// Package main demonstrates managing the BMC HTTPS certificate with bmclib: listing the installed
// certificates, generating a certificate signing request and installing a signed certificate.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Command line option flag parsing
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Password to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	csr := flag.Bool("csr", false, "Generate a certificate signing request for the BMC HTTPS endpoint and print it")
	commonName := flag.String("cn", "", "CSR subject common name (defaults to the BMC hostname)")
	organization := flag.String("o", "", "CSR subject organization")
	organizationalUnit := flag.String("ou", "", "CSR subject organizational unit")
	city := flag.String("l", "", "CSR subject locality (city)")
	state := flag.String("st", "", "CSR subject state or province")
	country := flag.String("c", "", "CSR subject country (two letters)")
	replace := flag.String("replace", "", "Path of a PEM-encoded certificate to install for the BMC HTTPS endpoint")

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

	// Filter to providers that support manager HTTPS certificate management
	client.Registry.Drivers = client.Registry.Supports(
		providers.FeatureGetManagerHTTPSCertificates,
		providers.FeatureGenerateManagerHTTPSCSR,
		providers.FeatureReplaceManagerHTTPSCertificate,
	)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	switch {
	case *csr:
		if *commonName == "" {
			*commonName = *host
		}
		pem, err := client.GenerateManagerHTTPSCSR(ctx, &bmc.CSRRequest{
			CommonName:         *commonName,
			Organization:       *organization,
			OrganizationalUnit: *organizationalUnit,
			City:               *city,
			State:              *state,
			Country:            *country,
		})
		if err != nil {
			l.Fatal(err)
		}
		fmt.Print(pem)
		if pem != "" && pem[len(pem)-1] != '\n' {
			fmt.Println()
		}
	case *replace != "":
		pem, err := os.ReadFile(*replace)
		if err != nil {
			l.Fatal(err)
		}
		if err := client.ReplaceManagerHTTPSCertificate(ctx, string(pem)); err != nil {
			l.Fatal(err)
		}
		fmt.Println("certificate replaced; the BMC web server may restart to apply it")
	}

	certificates, err := client.GetManagerHTTPSCertificates(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("HTTPS certificates:")
	for _, c := range certificates {
		fmt.Printf("  %-24s subject=%s issuer=%s valid=%s..%s serial=%s\n",
			c.ID, c.SubjectCommonName, c.IssuerCommonName, c.ValidNotBefore, c.ValidNotAfter, c.SerialNumber)
	}
}
