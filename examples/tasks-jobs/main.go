// Package main demonstrates listing the BMC task and job collections with bmclib.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
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

	// Filter to providers that support listing tasks and jobs
	client.Registry.Drivers = client.Registry.Supports(providers.FeatureGetTasks, providers.FeatureGetJobs)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	tasks, err := client.GetTasks(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Tasks:")
	for i := range tasks {
		t := &tasks[i]
		progress := "n/a"
		if t.PercentComplete != nil {
			progress = fmt.Sprintf("%d%%", *t.PercentComplete)
		}
		fmt.Printf("  %-20s %-32s %-12s %-10s %-5s start=%s %s\n",
			t.ID, t.Name, t.TaskState, t.TaskStatus, progress, t.StartTime, strings.Join(t.Messages, "; "))
	}

	jobs, err := client.GetJobs(ctx)
	if err != nil {
		if errors.Is(err, bmclibErrs.ErrJobServiceNotFound) {
			fmt.Println("Jobs: the BMC has no JobService")
			return
		}
		l.Fatal(err)
	}

	fmt.Println("Jobs:")
	for i := range jobs {
		j := &jobs[i]
		progress := "n/a"
		if j.PercentComplete != nil {
			progress = fmt.Sprintf("%d%%", *j.PercentComplete)
		}
		fmt.Printf("  %-20s %-32s %-12s %-10s %-5s start=%s %s\n",
			j.ID, j.Name, j.JobState, j.JobStatus, progress, j.StartTime, strings.Join(j.Messages, "; "))
	}
}
