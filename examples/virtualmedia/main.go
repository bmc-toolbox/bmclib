package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/metal-toolbox/bmclib"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	user := flag.String("user", "", "BMC username, required")
	pass := flag.String("password", "", "BMC password, required")
	host := flag.String("host", "", "BMC hostname or IP address, required")
	isoURL := flag.String("iso", "", "The HTTP URL to the ISO to be mounted, leave empty to unmount")
	flag.Parse()

	if *user == "" || *pass == "" || *host == "" {
		fmt.Fprintln(os.Stderr, "user, password, and host are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	log := logr.FromSlogHandler(l.Handler())

	cl := bmclib.NewClient(*host, *user, *pass, bmclib.WithLogger(log))
	if err := cl.Open(ctx); err != nil {
		panic(err)
	}
	defer cl.Close(ctx)

	ok, err := cl.SetVirtualMedia(ctx, "CD", *isoURL)
	if err != nil {
		log.Info("debugging", "metadata", cl.GetMetadata())
		panic(err)
	}
	if !ok {
		log.Info("debugging", "metadata", cl.GetMetadata())
		panic("failed virtual media operation")
	}
	log.Info("virtual media operation successful", "metadata", cl.GetMetadata())
}
