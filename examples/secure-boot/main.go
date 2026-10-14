package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
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
	mode := flag.String("mode", "get", "Mode [get,set,reset,reset-db,import]")
	database := flag.String("database", "db", "Database for database-specific operations: db, KEK, PK, dbx")
	certFile := flag.String("cert", "", "Path to PEM-encoded certificate file (for import mode)")

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

	// Filter to providers that support Secure Boot operations
	client.Registry.Drivers = client.Registry.Supports(
		providers.FeatureGetSecureBoot,
		providers.FeatureSetSecureBoot,
		providers.FeatureResetSecureBootKeys,
		providers.FeatureResetSecureBootDatabaseKeys,
		providers.FeatureImportSecureBootCertificate,
	)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	// Operating mode selection
	switch strings.ToLower(*mode) {
	case "get":
		// Get current Secure Boot state
		enabled, err := client.GetSecureBoot(ctx)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("Secure Boot enabled: %v\n", enabled)

	case "set":
		// Enable Secure Boot
		// To disable, use 'false' instead
		err := client.SetSecureBoot(ctx, true)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Println("Secure Boot enabled successfully")

	case "reset":
		// Reset all Secure Boot keys to default
		err := client.ResetSecureBootKeys(ctx, bmc.ResetSecureBootKeysTypeResetAllKeysToDefault)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Println("Secure Boot keys reset to default successfully")

	case "reset-db":
		// Reset a specific Secure Boot database
		// Valid databases: db, KEK, PK, dbx
		dbType := stringToSecureBootDatabase(*database)
		if dbType == "" {
			l.Fatalf("invalid database: %s (must be db, KEK, PK, or dbx)", *database)
		}

		err := client.ResetSecureBootDatabaseKeys(ctx, dbType, bmc.ResetSecureBootDatabaseKeysTypeResetAllKeysToDefault)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("Secure Boot database '%s' reset to default successfully\n", *database)

	case "import":
		// Import a certificate into a Secure Boot database
		if *certFile == "" {
			l.Fatal("certificate file required for import mode")
		}

		certData, err := os.ReadFile(*certFile)
		if err != nil {
			l.Fatal(err)
		}

		dbType := stringToSecureBootDatabase(*database)
		if dbType == "" {
			l.Fatalf("invalid database: %s (must be db or KEK)", *database)
		}

		err = client.ImportSecureBootCertificate(ctx, dbType, string(certData))
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("Certificate imported to database '%s' successfully\n", *database)

	default:
		l.Fatalf("unknown mode: %s", *mode)
	}
}

// stringToSecureBootDatabase converts a string to a SecureBootDatabase constant
func stringToSecureBootDatabase(s string) bmc.SecureBootDatabase {
	switch strings.ToLower(s) {
	case "db":
		return bmc.SecureBootDatabaseDB
	case "kek":
		return bmc.SecureBootDatabaseKEK
	case "pk":
		return bmc.SecureBootDatabasePK
	case "dbx":
		return bmc.SecureBootDatabaseDBX
	default:
		return ""
	}
}
