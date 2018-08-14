package butler

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/asset"

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

	//time how long it takes to setup a connection
	metricPrefix := fmt.Sprintf("%s.%s.%s", asset.Location, asset.Vendor, asset.Type)
	defer b.metricsEmitter.MeasureRunTime(
		time.Now().Unix(), fmt.Sprintf("%s.%s", metricPrefix, component))

	bmcPrimaryUser := viper.GetString("bmcPrimaryUser")
	bmcPrimaryPassword := viper.GetString("bmcPrimaryPassword")

	client, err := discover.ScanAndConnect(asset.IpAddress, bmcPrimaryUser, bmcPrimaryPassword)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"IP":        asset.IpAddress,
			"butler-id": b.id,
			"Error":     err,
		}).Warn("Unable to connect to bmc.")
		return connection, err
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)
		asset.Model = bmc.BmcType()

		//we don't check credentials if its an ssh based connection
		if !dontCheckCredentials {

			//attempt to login with Primary user account
			err := bmc.CheckCredentials()
			if err == bmcerros.ErrLoginFailed {
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": bmcPrimaryUser,
					"Error":        err,
				}).Warn("Unable to login to bmc with Primary user, trying Secondary/Default user account.")

				//attempt to login with Secondary user account
				bmcSecondaryUser := viper.GetString("bmcSecondaryUser")
				if bmcSecondaryUser != "" {
					bmcSecondaryPassword := viper.GetString("bmcSecondaryPassword")
					bmc.UpdateCredentials(bmcSecondaryUser, bmcSecondaryPassword)

					err := bmc.CheckCredentials()
					if err != nil {
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": bmcSecondaryUser,
							"Error":          err,
						}).Warn("Unable to login to bmc with Secondary user, will attempt to login with vendor default credentials.")
						return bmc, err
					}

					//successful login with Secondary user
					log.WithFields(logrus.Fields{
						"component":      component,
						"butler-id":      b.id,
						"Asset":          fmt.Sprintf("%+v", asset),
						"Secondary user": bmcSecondaryUser,
					}).Debug("Successful login with Secondary user.")

					asset.Vendor = bmc.Vendor()
					return bmc, err
				}

				//attempt to login with vendor Default user account
				bmcDefaultUser := viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
				bmcDefaultPassword := viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
				bmc.UpdateCredentials(bmcDefaultUser, bmcDefaultPassword)
				err := bmc.CheckCredentials()
				if err != nil {
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": bmcDefaultUser,
						"Error":        err,
					}).Warn("Unable to login to bmc with default credentials.")
					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": bmcDefaultUser,
					}).Debug("Successful login with vendor default user.")

					asset.Vendor = bmc.Vendor()
					return bmc, err

				}
			} else if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Unable to connect to BMC.")
				return bmc, err
			} else {

				//successful login - Primary user
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": bmcPrimaryUser,
				}).Debug("Successful login with Primary user.")
				asset.Vendor = bmc.Vendor()
				return bmc, err
			}
		}

		return bmc, err

	case devices.BmcChassis:
		bmc := client.(devices.BmcChassis)
		asset.Model = bmc.BmcType()

		//we don't check credentials if its an ssh based connection
		if !dontCheckCredentials {

			//attempt to login with Primary user account
			err := bmc.CheckCredentials()
			if err == bmcerros.ErrLoginFailed {
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": bmcPrimaryUser,
					"Error":        err,
				}).Warn("Unable to login to bmc with Primary user, trying Secondary/Default user account.")

				//attempt to login with Secondary user account
				bmcSecondaryUser := viper.GetString("bmcSecondaryUser")
				if bmcSecondaryUser != "" {
					bmcSecondaryPassword := viper.GetString("bmcSecondaryPassword")
					bmc.UpdateCredentials(bmcSecondaryUser, bmcSecondaryPassword)

					err := bmc.CheckCredentials()
					if err != nil {
						log.WithFields(logrus.Fields{
							"component":      component,
							"butler-id":      b.id,
							"Asset":          fmt.Sprintf("%+v", asset),
							"Secondary user": bmcSecondaryUser,
							"Error":          err,
						}).Warn("Unable to login to bmc with Secondary user, will attempt to login with vendor default credentials.")
						return bmc, err
					}

					//successful login with Secondary user
					log.WithFields(logrus.Fields{
						"component":      component,
						"butler-id":      b.id,
						"Asset":          fmt.Sprintf("%+v", asset),
						"Secondary user": bmcSecondaryUser,
					}).Debug("Successful login with Secondary user.")

					asset.Vendor = bmc.Vendor()
					return bmc, err
				}

				//attempt to login with vendor Default user account
				bmcDefaultUser := viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
				bmcDefaultPassword := viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
				bmc.UpdateCredentials(bmcDefaultUser, bmcDefaultPassword)
				err := bmc.CheckCredentials()
				if err != nil {
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": bmcDefaultUser,
						"Error":        err,
					}).Warn("Unable to login to bmc with default credentials.")
					return bmc, err
				} else {

					//successful login - with default credentials
					log.WithFields(logrus.Fields{
						"component":    component,
						"butler-id":    b.id,
						"Asset":        fmt.Sprintf("%+v", asset),
						"Default user": bmcDefaultUser,
					}).Debug("Successful login with vendor default user.")

					asset.Vendor = bmc.Vendor()
					return bmc, err

				}
			} else if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Unable to connect to BMC.")
				return bmc, err
			} else {

				//successful login - Primary user
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    b.id,
					"Asset":        fmt.Sprintf("%+v", asset),
					"Primary user": bmcPrimaryUser,
				}).Debug("Successful login with Primary user.")
				asset.Vendor = bmc.Vendor()
				return bmc, err
			}
		}
	}

	return connection, err
}
