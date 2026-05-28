package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	user := flag.String("user", "", "BMC username, required")
	pass := flag.String("password", "", "BMC password, required")
	host := flag.String("host", "", "BMC hostname or IP address, required")
	isoURL := flag.String("iso", "", "The HTTP URL to the ISO/image to be mounted")
	mediaKind := flag.String("media-kind", "CD", "Virtual media kind: CD, DVD, Floppy, USBStick")
	eject := flag.Bool("eject", false, "Eject/unmount virtual media instead of mounting")
	flag.Parse()

	if *user == "" || *pass == "" || *host == "" {
		fmt.Fprintln(os.Stderr, "user, password, and host are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !*eject && *isoURL == "" {
		fmt.Fprintln(os.Stderr, "iso is required unless -eject is set")
		flag.PrintDefaults()
		os.Exit(1)
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	log := logr.FromSlogHandler(l.Handler())

	cl := bmclib.NewClient(*host, *user, *pass, bmclib.WithLogger(log))
	cl.Registry.Drivers = cl.Registry.For("gofish")
	if err := cl.Open(ctx); err != nil {
		panic(err)
	}
	defer cl.Close(ctx)

	mediaURL := *isoURL
	operation := "mount"
	if *eject {
		mediaURL = ""
		operation = "eject"
	}

	ok, err := cl.SetVirtualMedia(ctx, *mediaKind, mediaURL)
	if err != nil {
		log.Info("debugging", "metadata", cl.GetMetadata())
		panic(err)
	}
	if !ok {
		log.Info("debugging", "metadata", cl.GetMetadata())
		panic("failed virtual media operation")
	}
	log.Info(
		"virtual media operation successful",
		"operation", operation,
		"media-kind", *mediaKind,
		"metadata", cl.GetMetadata(),
	)
}
