// Package main demonstrates reading the chassis temperature sensors and fans with bmclib.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Command line option flag parsing
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Password to login with")
	host := flag.String("host", "", "BMC hostname to connect to")

	flag.Parse()

	// Logger configuration
	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	// Validate required parameters
	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	// bmclib client abstraction
	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}
	client := bmclib.NewClient(*host, *user, *pass, clientOpts...)

	// Filter to providers that support chassis fan and temperature readings
	client.Registry.Drivers = client.Registry.Supports(providers.FeatureGetChassisTemperatures, providers.FeatureGetChassisFanSpeeds)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	temperatures, err := client.GetChassisTemperatures(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Temperatures:")
	for _, t := range temperatures {
		reading := "n/a"
		if t.ReadingCelsius != nil {
			reading = fmt.Sprintf("%.0f C", *t.ReadingCelsius)
		}
		fmt.Printf("  %-30s %-8s health=%s state=%s\n", t.Name, reading, t.Health, t.State)
	}

	fans, err := client.GetChassisFanSpeeds(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Fans:")
	for _, f := range fans {
		speed := "n/a"
		switch {
		case f.SpeedRPM != nil && f.SpeedPercent != nil:
			speed = fmt.Sprintf("%.0f RPM (%.0f%%)", *f.SpeedRPM, *f.SpeedPercent)
		case f.SpeedRPM != nil:
			speed = fmt.Sprintf("%.0f RPM", *f.SpeedRPM)
		case f.SpeedPercent != nil:
			speed = fmt.Sprintf("%.0f%%", *f.SpeedPercent)
		}
		fmt.Printf("  %-30s %-16s health=%s state=%s\n", f.Name, speed, f.Health, f.State)
	}
}
