/*
Package main provides an example of using the bmclib Secure Boot interfaces
to get/set UEFI Secure Boot state and manage Secure Boot key databases.

This example demonstrates:
  - Getting the current Secure Boot state
  - Enabling/disabling Secure Boot
  - Resetting all Secure Boot keys
  - Resetting a specific Secure Boot database
  - Importing a certificate into a Secure Boot database

Usage:

	$ go run ./examples/secure-boot/main.go -h
	Usage of ./main:
	  -host string
	        BMC hostname to connect to
	  -password string
	        Password to login with
	  -user string
	        Username to login with
	  -cert string
	        Path to PEM-encoded certificate file (for import mode)
	  -database string
	        Database for database-specific operations: db, KEK, PK, dbx (default "db")
	  -mode string
	        Mode [get,set,reset,reset-db,import] (default "get")
*/
package main
