package butler

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerros "github.com/bmc-toolbox/bmclib/errors"
)

// Sets up the connection to the asset
// Attempts login with current, if that fails tries with default passwords.
// Returns a connection interface that can be type cast to devices.Bmc or devices.BmcChassis
func (b *Butler) setupConnection(asset *asset.Asset, dontCheckCredentials bool) (connection interface{}, err error) {

	log := b.log
	component := "setupConnection"
	metric := b.metricsEmitter

	var user, password string
	defer metric.MeasureRuntime([]string{"butler", "conn_setup_runtime"}, time.Now())

	user = b.config.BmcPrimaryUser
	password = b.config.BmcPrimaryPassword

	client, err := discover.ScanAndConnect(asset.IpAddress, user, password)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"IP":        asset.IpAddress,
			"butler-id": b.id,
			"Error":     err,
		}).Warn("Error connecting to bmc.")

		defer metric.IncrCounter([]string{"butler", "conn_setup_failed"}, 1)
		return connection, err
	}

	//connect success
	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)
		asset.Vendor = bmc.Vendor()

		//we don't check credentials if its an ssh based connection
		if !dontCheckCredentials {

			//attempt to login with Primary user account
			err := bmc.CheckCredentials()
			if err != nil {
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": user,
					"Error":        err,
				}).Warn("Login with Primary user failed.")

				metric.IncrCounter(
					[]string{
						"butler",
						fmt.Sprintf("user_login_failed_%s", user),
					}, 1)

				//attempt to login with Secondary user account
				user = b.config.BmcSecondaryUser
				if user != "" {
					password = b.config.BmcSecondaryPassword
					bmc.UpdateCredentials(user, password)

					err := bmc.CheckCredentials()
					if err != nil {
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
							"Error":          err,
						}).Warn("Login with Secondary user failed.")

						metric.IncrCounter(
							[]string{
								"butler",
								fmt.Sprintf("user_login_failed_%s", user),
							}, 1)
					} else {

						//successful login with Secondary user
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
						}).Debug("Successful login with Secondary user.")

						asset.Vendor = bmc.Vendor()

						metric.IncrCounter([]string{
							"butler",
							fmt.Sprintf("user_login_success_%s", user),
						}, 1)
						return bmc, err
					}
				}

				//read in vendor default credentials and attempt login.
				err = b.config.GetDefaultCredentials(asset.Vendor)
				if err != nil {
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": b.id,
						"Asset":     fmt.Sprintf("%+v", asset),
						"Error":     err,
					}).Warn("No vendor default user creds found in config.")

					metric.IncrCounter(
						[]string{
							"butler",
							"login_failed_no_default_creds",
						}, 1)
					return bmc, err
				}

				user = b.config.BmcDefaultUser
				password = b.config.BmcDefaultPassword

				bmc.UpdateCredentials(user, password)
				err := bmc.CheckCredentials()
				if err != nil {
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
						"Error":        err,
					}).Warn("Login with default credentials failed.")

					metric.IncrCounter(
						[]string{"butler",
							fmt.Sprintf("user_login_failed_%s", user),
						}, 1)
					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
					}).Debug("Successful login with vendor default user.")
					metric.IncrCounter([]string{
						"butler",
						fmt.Sprintf("user_login_success_%s", user),
					}, 1)

					return bmc, err
				}
			} else {

				//successful login - Primary user
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": user,
				}).Debug("Successful login with Primary user.")
				metric.IncrCounter([]string{
					"butler",
					fmt.Sprintf("user_login_success_%s", user),
				}, 1)

				return bmc, err
			}
		}

		return bmc, err

	case devices.BmcChassis:
		bmc := client.(devices.BmcChassis)
		asset.Model = bmc.BmcType()
		asset.Vendor = bmc.Vendor()

		//we don't check credentials if its an ssh based connection
		if !dontCheckCredentials {

			//attempt to login with Primary user account
			err := bmc.CheckCredentials()

			//C7000's just timeout on auth failure.
			if err == bmcerros.ErrLoginFailed || (err != nil && asset.Model == "c7000") {

				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": user,
					"Error":        err,
				}).Warn("Login with Primary user failed.")

				metric.IncrCounter(
					[]string{
						"butler",
						fmt.Sprintf("user_login_failed_%s", user),
					}, 1)

				user = b.config.BmcSecondaryUser
				if user != "" {
					password = b.config.BmcSecondaryPassword
					bmc.UpdateCredentials(user, password)

					err := bmc.CheckCredentials()
					if err != nil {
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
							"Error":          err,
						}).Warn("Login with Secondary user failed.")

						metric.IncrCounter(
							[]string{
								"butler",
								fmt.Sprintf("user_login_failed_%s", user),
							}, 1)

					} else {

						//successful login with Secondary user
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
						}).Debug("Successful login with Secondary user.")

						asset.Vendor = bmc.Vendor()

						metric.IncrCounter([]string{
							"butler",
							fmt.Sprintf("user_login_success_%s", user),
						}, 1)
						return bmc, err
					}
				}
				//read in vendor default credentials and attempt login.
				err = b.config.GetDefaultCredentials(asset.Vendor)
				if err != nil {
					return bmc, err
				}

				user = b.config.BmcDefaultUser
				password = b.config.BmcDefaultPassword

				bmc.UpdateCredentials(user, password)
				err := bmc.CheckCredentials()
				if err != nil {
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
						"Error":        err,
					}).Warn("Login with default credentials failed.")

					metric.IncrCounter(
						[]string{
							"butler",
							fmt.Sprintf("user_login_failed_%s", user),
						}, 1)

					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
					}).Debug("Successful login with vendor default user.")

					metric.IncrCounter([]string{
						"butler",
						fmt.Sprintf("user_login_success_%s", user),
					}, 1)
					return bmc, err
				}
			} else if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("BMC connection failed.")

				defer metric.IncrCounter([]string{"butler", "conn_setup_failed"}, 1)
				return bmc, err
			} else {

				//successful login - Primary user
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": user,
				}).Debug("Successful login with Primary user.")

				metric.IncrCounter([]string{
					"butler",
					fmt.Sprintf("user_login_success_%s", user),
				}, 1)
				return bmc, err
			}
		}
	}

	return connection, err
}
