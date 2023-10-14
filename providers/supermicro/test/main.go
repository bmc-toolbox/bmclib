package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := c.Open(ctx); err != nil {
		log.Fatal("login error" + err.Error())
	}

	defer c.Close(ctx)

	// open file handle
	fh, err := os.Open("BMC_X12AST2600-F201MS_20220627_1.13.04_STDsp.bin")
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	uploadTaskID, err := c.FirmwareUpload(ctx, common.SlugBMC, constants.OnStartUpdateRequest, fh)
	if err != nil {
		l.Fatal(err)
	}

	opts := &bmc.FirmwareInstallOptions{
		UploadTaskID: uploadTaskID,
	}

	taskID, err := c.FirmwareInstallWithOptions(ctx, common.SlugBMC, nil, opts)
	if err != nil {
		l.Fatal(err)
	}

	log.Println(taskID)
}
