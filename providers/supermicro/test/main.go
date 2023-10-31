package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
	brrs "github.com/bmc-toolbox/bmclib/v2/errors"
	smc "github.com/bmc-toolbox/bmclib/v2/providers/supermicro"
	"github.com/bmc-toolbox/common"
	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)
	c := smc.NewClient("10.251.153.157", "ADMIN", "XWMCYBJEPL", logger)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	if err := c.Open(ctx); err != nil {
		log.Fatal("login error" + err.Error())
	}

	defer c.Close(ctx)

	// open file handle
	//f := "BMC_X12AST2600-F201MS_20220627_1.13.04_STDsp.bin"
	f := "BIOS_X12STH-1C3A_20230607_1.6_STDsp.bin"
	fh, err := os.Open(f)
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	//	component := common.SlugBMC
	component := common.SlugBIOS

	uploadTaskID, err := c.FirmwareUpload(ctx, component, fh)
	if err != nil {
		if errors.Is(err, brrs.ErrBMCColdResetRequired) {
			log.Println(err)

			if err := c.ResetBMC(ctx); err != nil {
				l.Fatal(err)
			}

			log.Fatal("bmc reset, retry in a few minutes")
		}

		l.Fatal(err)
	}

	log.Println("upload taskID: " + uploadTaskID)

	opts := &bmc.FirmwareInstallOptions{
		UploadTaskID: uploadTaskID,
	}

	var installTaskID string
	for {
		installTaskID, err = c.FirmwareInstallWithOptions(ctx, component, nil, opts)
		if err != nil {
			if errors.Is(err, brrs.ErrFirmwareVerifyTaskRunning) {
				log.Println("retrying in 5 secs..: ", err.Error())

				time.Sleep(5 * time.Second)
				continue

			}

			l.Fatal(err)
		} else {
			log.Println("install task in progress: ", installTaskID)
			break
		}
	}

	for {
		if ctx.Err() != nil {
			l.Fatal(ctx.Err())
		}

		state, err := c.FirmwareInstallStatus(ctx, "", component, installTaskID)
		if err != nil {
			// when its under update a connection refused is returned
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "operation timed out") {
				l.Info("BMC refused connection, BMC most likely resetting...")
				time.Sleep(2 * time.Second)

				continue
			}

			if errors.Is(err, brrs.ErrSessionExpired) || strings.Contains(err.Error(), "session expired") {
				err := c.Open(ctx)
				if err != nil {
					l.Fatal(err, "bmc re-login failed")
				}

				l.WithFields(logrus.Fields{"state": state, "component": component}).Info("BMC session expired, logging in...")

				continue
			}

			log.Fatal(err)
		}

		switch state {
		case constants.FirmwareInstallRunning, constants.FirmwareInstallInitializing:
			l.WithFields(logrus.Fields{"state": state, "component": component}).Info("firmware install running")

		case constants.FirmwareInstallFailed:
			l.WithFields(logrus.Fields{"state": state, "component": component}).Info("firmware install failed")
			os.Exit(1)

		case constants.FirmwareInstallComplete:
			l.WithFields(logrus.Fields{"state": state, "component": component}).Info("firmware install completed")
			os.Exit(0)

		case constants.FirmwareInstallPowerCyleHost:
			l.WithFields(logrus.Fields{"state": state, "component": component}).Info("host powercycle required")

			//	if _, err := c.SetPowerState(ctx, "cycle"); err != nil {
			//		l.WithFields(logrus.Fields{"state": state, "component": *component}).Info("error power cycling host for install")
			os.Exit(1)
			////	}
			//
			//l.WithFields(logrus.Fields{"state": state, "component": *component}).Info("host power cycled, all done!")
			//os.Exit(0)
		default:
			l.WithFields(logrus.Fields{"state": state, "component": component}).Info("unknown state returned")
		}

		time.Sleep(2 * time.Second)
	}

}
