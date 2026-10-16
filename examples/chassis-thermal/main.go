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

	// Filter to providers that support chassis thermal readings
	client.Registry.Drivers = client.Registry.Supports(providers.FeatureGetChassisThermal)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	thermal, err := client.GetChassisThermal(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Temperatures:")
	for _, t := range thermal.Temperatures {
		reading := "n/a"
		if t.ReadingCelsius != nil {
			reading = fmt.Sprintf("%.0f C", *t.ReadingCelsius)
		}
		fmt.Printf("  %-30s %-8s health=%s state=%s\n", t.Name, reading, t.Health, t.State)
	}

	fmt.Println("Fans:")
	for _, f := range thermal.Fans {
		reading := "n/a"
		if f.Reading != nil {
			reading = fmt.Sprintf("%d %s", *f.Reading, f.ReadingUnits)
		}
		fmt.Printf("  %-30s %-10s health=%s state=%s\n", f.Name, reading, f.Health, f.State)
	}
}
