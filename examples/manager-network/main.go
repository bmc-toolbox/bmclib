// Package main demonstrates reading the BMC network configuration with bmclib: the manager
// ethernet interfaces and the management services (protocols) the BMC reports.
//
// With -set-service or -ntp-servers it first changes the manager network protocol resource:
//
//	manager-network -host ... -set-service IPMI -enabled=false
//	manager-network -host ... -set-service SSH -enabled=true -port 2222
//	manager-network -host ... -ntp-servers 10.0.0.1,10.0.0.2
//
// With -interface and -set-hostname or -set-name-servers it changes that manager ethernet interface:
//
//	manager-network -host ... -interface NIC.1 -set-hostname bmc-r12-u07
//	manager-network -host ... -interface NIC.1 -set-name-servers 10.0.0.2,2001:db8::53
//
// Disabling the service this example connects through (HTTPS) cuts it off from the BMC.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/netip"
	"strings"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/jacobweinstock/registrar"
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
	setService := flag.String("set-service", "", "Service to enable or disable (HTTP, HTTPS, SSH, Telnet, IPMI, SNMP, NTP, KVMIP, VirtualMedia, SSDP)")
	enabled := flag.Bool("enabled", true, "With -set-service: the enabled state to set")
	port := flag.Int("port", 0, "With -set-service: the port to set; 0 leaves the port as is")
	ntpServers := flag.String("ntp-servers", "", "Comma separated NTP servers to set; an empty value clears the list")
	iface := flag.String("interface", "", "With -set-hostname or -set-name-servers: ID of the manager ethernet interface to change, e.g. NIC.1 or eth0")
	setHostName := flag.String("set-hostname", "", "Host name to set on the manager ethernet interface given with -interface")
	nameServers := flag.String("set-name-servers", "", "Comma separated static DNS server addresses to set on the interface given with -interface; an empty value clears the list")

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

	// -ntp-servers "" and -set-name-servers "" clear the list, so the flag being given matters,
	// not its value
	setNTPServers, setNameServers := false, false
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "ntp-servers":
			setNTPServers = true
		case "set-name-servers":
			setNameServers = true
		}
	})

	if (*setHostName != "" || setNameServers) && *iface == "" {
		l.Fatal("-set-hostname and -set-name-servers require -interface")
	}

	// The name servers are typed addresses, so a host name is refused here, before the BMC is contacted
	var staticNameServers []netip.Addr
	if *nameServers != "" {
		for _, server := range strings.Split(*nameServers, ",") {
			addr, err := netip.ParseAddr(strings.TrimSpace(server))
			if err != nil {
				l.Fatal(err)
			}
			staticNameServers = append(staticNameServers, addr)
		}
	}

	// Filter to providers that can read, and when asked change, the manager network configuration
	features := []registrar.Feature{
		providers.FeatureGetManagerEthernetInterfaces,
		providers.FeatureGetManagerNetworkProtocol,
	}
	if *setService != "" {
		features = append(features, providers.FeatureSetManagerNetworkService)
	}
	if setNTPServers {
		features = append(features, providers.FeatureSetManagerNTPServers)
	}
	if *setHostName != "" {
		features = append(features, providers.FeatureSetManagerEthernetInterfaceHostName)
	}
	if setNameServers {
		features = append(features, providers.FeatureSetManagerEthernetInterfaceNameServers)
	}
	client.Registry.Drivers = client.Registry.Supports(features...)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	if *setService != "" {
		var servicePort *int
		if *port != 0 {
			servicePort = port
		}
		if err := client.SetManagerNetworkService(ctx, bmc.ManagerNetworkServiceName(*setService), *enabled, servicePort); err != nil {
			l.Fatal(err)
		}
		fmt.Printf("Service %s: enabled=%t port=%s requested\n", *setService, *enabled, intOrDash(servicePort))
	}

	if setNTPServers {
		var servers []string
		if *ntpServers != "" {
			servers = strings.Split(*ntpServers, ",")
		}
		if err := client.SetManagerNTPServers(ctx, servers); err != nil {
			l.Fatal(err)
		}
		fmt.Printf("NTP servers: %s requested\n", strings.Join(servers, ", "))
	}

	if *setHostName != "" {
		if err := client.SetManagerEthernetInterfaceHostName(ctx, *iface, *setHostName); err != nil {
			l.Fatal(err)
		}
		fmt.Printf("Interface %s: host name %s requested\n", *iface, *setHostName)
	}

	if setNameServers {
		if err := client.SetManagerEthernetInterfaceNameServers(ctx, *iface, staticNameServers); err != nil {
			l.Fatal(err)
		}
		fmt.Printf("Interface %s: name servers %v requested\n", *iface, staticNameServers)
	}

	// A BMC may apply the change after answering the request, so what is read back below can
	// still show the previous state.
	interfaces, err := client.GetManagerEthernetInterfaces(ctx)
	if err != nil {
		l.Fatal(err)
	}

	fmt.Println("Manager ethernet interfaces:")
	for i := range interfaces {
		eth := &interfaces[i]
		fmt.Printf("  %-8s hostname=%s mac=%s enabled=%t link=%s speed=%s mtu=%s vlan=%s\n",
			eth.ID, eth.HostName, eth.MACAddress, eth.Enabled, eth.LinkStatus, intOrDash(eth.SpeedMbps), intOrDash(eth.MTUSize), vlan(eth))
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
