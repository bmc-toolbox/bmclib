package main

import (
	"context"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bombsimon/logrusr/v2"
	bmclib "github.com/metal-toolbox/bmclib"
	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	component := flag.String("component", "", "Component to be updated (bmc, bios.. etc)")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolPath := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	firmwarePath := flag.String("firmware", "", "The local path of the firmware to install")
	firmwareVersion := flag.String("version", "", "The firmware version being installed")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	l := logrus.New()
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)

	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	if *component == "" {
		l.Fatal("component parameter required (must be a component slug - bmc, bios etc)")
	}

	clientOpts := []bmclib.Option{
		bmclib.WithLogger(logger),
		bmclib.WithPerProviderTimeout(time.Minute * 30),
	}

	if *withSecureTLS {
		var pool *x509.CertPool
		if *certPoolPath != "" {
			pool = x509.NewCertPool()
			data, err := ioutil.ReadFile(*certPoolPath)
			if err != nil {
				l.Fatal(err)
			}
			pool.AppendCertsFromPEM(data)
		}
		// a nil pool uses the system certs
		clientOpts = append(clientOpts, bmclib.WithSecureTLS(pool))
	}

	cl := bmclib.NewClient(*host, *user, *pass, clientOpts...)

	err := cl.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	// open file handle
	fh, err := os.Open(*firmwarePath)
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	steps, err := cl.FirmwareInstallSteps(ctx, *component)
	if err != nil {
		l.Fatal(err)
	}

	sprinted := fmt.Sprintf("%v", steps)
	trimmed := strings.Trim(sprinted, "[]")
	replaced := strings.Replace(trimmed, " ", " - ", 0)

	l.Infof("Steps: %s", replaced)

	taskID := ""
	var lastStep constants.FirmwareInstallStep = ""
	for _, step := range steps {
		l.Infof("Step: %s", step)

		switch step {
		case constants.FirmwareInstallStepUploadInitiateInstall:
			taskID, err = cl.FirmwareInstallUploadAndInitiate(ctx, *component, fh)
			if err != nil {
				l.Fatal(err)
			}
			// X11 doesnt have a taskID, so lets give it a dummy one
			if taskID == "" {
				taskID = "0"
			}
		case constants.FirmwareInstallStepInstallStatus:
			fallthrough
		case constants.FirmwareInstallStepUploadStatus:
			if taskID == "" {
				l.Warn("taskID wasnt set, continueing anyway")
			}
			if lastStep == "" {
				l.Fatal("lastStep wasnt set")
			}
			firmwareInstallStatusWait(ctx, cl, l, lastStep, *component, *firmwareVersion, taskID)
		case constants.FirmwareInstallStepUpload:
			taskID, err = cl.FirmwareUpload(ctx, *component, fh)
			if err != nil {
				l.Fatal(err)
			}
			// X11 doesnt have a taskID, so lets give it a dummy one
			if taskID == "" {
				taskID = "0"
			}
		case constants.FirmwareInstallStepInstallUploaded:
			if taskID == "" {
				l.Fatal("taskID wasnt set")
			}
			taskID, err = cl.FirmwareInstallUploaded(ctx, *component, taskID)
			if err != nil {
				l.Fatal(err)
			}
			// X11 doesnt have a taskID, so lets give it a dummy one
			if taskID == "" {
				taskID = "0"
			}
		case constants.FirmwareInstallStepPowerOffHost:
			_, err = cl.SetPowerState(ctx, "off")
			if err != nil {
				l.Fatal(err)
			}
		case constants.FirmwareInstallStepResetBMCPostInstall:
			fallthrough
		case constants.FirmwareInstallStepResetBMCOnInstallFailure:
			_, err = cl.ResetBMC(ctx, "GracefulRestart")
			if err != nil {
				l.Fatal(err)
			}
		default:
			l.Fatal("unknown firmware install step")
		}

		lastStep = step
	}
}

func firmwareInstallStatusWait(ctx context.Context, cl *bmclib.Client, l *logrus.Logger, step constants.FirmwareInstallStep, component, firmwareVersion, taskID string) {
	for range 300 {
		if ctx.Err() != nil {
			l.Fatal(ctx.Err())
		}

		state, status, err := cl.FirmwareTaskStatus(ctx, step, component, taskID, firmwareVersion)
		if err != nil {
			// when its under update a connection refused is returned
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "operation timed out") {
				l.Info("BMC refused connection, BMC most likely resetting...")
				time.Sleep(2 * time.Second)

				continue
			}

			if errors.Is(err, bmclibErrs.ErrSessionExpired) || strings.Contains(err.Error(), "session expired") {
				err := cl.Open(ctx)
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
			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Infof("%s running", step)
		case constants.FirmwareInstallFailed:
			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Infof("%s failed", step)
			os.Exit(1)
		case constants.FirmwareInstallComplete:
			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Infof("%s completed", step)
			return
		case constants.FirmwareInstallPowerCycleHost:
			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Info("host powercycle required")

			if _, err := cl.SetPowerState(ctx, "cycle"); err != nil {
				l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Infof("error power cycling host for %s", step)
				os.Exit(1)
			}

			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Info("host power cycled, all done!")
			return
		default:
			l.WithFields(logrus.Fields{"state": state, "status": status, "component": component}).Info("unknown state returned")
		}

		time.Sleep(2 * time.Second)
	}
}
