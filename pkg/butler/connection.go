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

	var user, password string
	var pUserAuthSuccessMetric, pUserAuthFailMetric, sUserAuthSuccessMetric, sUserAuthFailMetric string
	var dUserAuthSuccessMetric, dUserAuthFailMetric, connfailMetric string

	//Setup a few metrics we will send out.
	metricPrefix := fmt.Sprintf("%s.%s.%s", asset.Location, asset.Vendor, asset.Type)

	pUserAuthSuccessMetric = fmt.Sprintf("%s.%s", metricPrefix, "primaryUserAuthSuccess")
	pUserAuthFailMetric = fmt.Sprintf("%s.%s", metricPrefix, "primaryUserAuthFail")

	sUserAuthSuccessMetric = fmt.Sprintf("%s.%s", metricPrefix, "secondaryUserAuthSuccess")
	sUserAuthFailMetric = fmt.Sprintf("%s.%s", metricPrefix, "secondaryUserAuthFail")

	dUserAuthSuccessMetric = fmt.Sprintf("%s.%s", metricPrefix, "defaultUserAuthSuccess")
	dUserAuthFailMetric = fmt.Sprintf("%s.%s", metricPrefix, "defaultUserAuthFail")

	connfailMetric = fmt.Sprintf("%s.%s", metricPrefix, "connfail")

	defer b.metricsEmitter.MeasureRunTime(
		time.Now().Unix(), fmt.Sprintf("%s.%s", metricPrefix, component))

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
		return connection, err
	}

	//auth success
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

				b.metricsData[pUserAuthFailMetric] += 1

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
						b.metricsData[sUserAuthFailMetric] += 1
					} else {

						//successful login with Secondary user
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
						}).Debug("Successful login with Secondary user.")

						asset.Vendor = bmc.Vendor()

						b.metricsData[sUserAuthSuccessMetric] += 1
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
					}).Warn("Unable to attempt login with vendor default user.")
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

					b.metricsData[dUserAuthFailMetric] += 1
					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
					}).Debug("Successful login with vendor default user.")

					b.metricsData[dUserAuthSuccessMetric] += 1
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

				b.metricsData[pUserAuthSuccessMetric] += 1
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

				b.metricsData[pUserAuthFailMetric] += 1

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
						b.metricsData[sUserAuthFailMetric] += 1
					} else {

						//successful login with Secondary user
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": user,
						}).Debug("Successful login with Secondary user.")

						asset.Vendor = bmc.Vendor()

						b.metricsData[sUserAuthSuccessMetric] += 1
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

					b.metricsData[dUserAuthFailMetric] += 1
					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": user,
					}).Debug("Successful login with vendor default user.")

					b.metricsData[dUserAuthSuccessMetric] += 1
					return bmc, err
				}
			} else if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("BMC connection failed.")

				b.metricsData[connfailMetric] += 1
				return bmc, err
			} else {

				//successful login - Primary user
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": user,
				}).Debug("Successful login with Primary user.")

				b.metricsData[pUserAuthSuccessMetric] += 1
				return bmc, err
			}
		}
	}

	return connection, err
}
