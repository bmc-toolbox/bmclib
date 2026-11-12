// Package main demonstrates reading the BMC network configuration with bmclib: the manager
// ethernet interfaces and the management services (protocols) the BMC reports.
package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	// Filter to providers that can read the manager network configuration
	client.Registry.Drivers = client.Registry.Supports(
		providers.FeatureGetManagerEthernetInterfaces,
		providers.FeatureGetManagerNetworkProtocol,
	)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	interfaces, err := client.GetManagerEthernetInterfaces(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Manager ethernet interfaces:")
	for i := range interfaces {
		eth := &interfaces[i]
		fmt.Printf("  %-8s mac=%s enabled=%t link=%s speed=%s mtu=%s vlan=%s\n",
			eth.ID, eth.MACAddress, eth.Enabled, eth.LinkStatus, intOrDash(eth.SpeedMbps), intOrDash(eth.MTUSize), vlan(eth))
		for _, a := range eth.IPv4Addresses {
			fmt.Printf("           ipv4 %s/%s gw=%s (%s)\n", a.Address, a.SubnetMask, a.Gateway, a.Origin)
		}
		for _, a := range eth.IPv6Addresses {
			fmt.Printf("           ipv6 %s/%d (%s)\n", a.Address, a.PrefixLength, a.Origin)
		}
		if len(eth.NameServers) > 0 {
			fmt.Printf("           dns  %s\n", strings.Join(eth.NameServers, ", "))
		}
	}

	protocol, err := client.GetManagerNetworkProtocol(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Printf("Manager network protocol: hostname=%s fqdn=%s ntp=%s\n", protocol.HostName, protocol.FQDN, strings.Join(protocol.NTPServers, ", "))
	for _, s := range protocol.Services {
		fmt.Printf("  %-13s enabled=%-5t port=%s\n", s.Name, s.Enabled, intOrDash(s.Port))
	}
}

func intOrDash(v *int) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprint(*v)
}

func vlan(eth *bmc.ManagerEthernetInterface) string {
	if !eth.VLANEnabled {
		return "off"
	}
	return fmt.Sprint(eth.VLANID)
}
