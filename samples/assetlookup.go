package main

// A sample assetlookup tool used by tests
// See docs/assetLookup.md for details.

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// AssetAttributes struct is the JSON format for asset data
type AssetAttributes struct {
	Data        map[string]Attributes `json:"data"` //map of asset IPs/Serials to attributes
	EndOfAssets bool                  `json:"end_of_assets,omitempty"`
	Offset      int                   `json:"offset,omitempty"`
	Limit       int                   `json:"limit,omitempty"`
}

// Attributes struct is the JSON format for asset attribute.
type Attributes struct {
	Serial    string           `json:"serial"`
	Location  string           `json:"location"`
	IPAddress []string         `json:"ipaddress"` //the BMC address.
	Extras    AttributesExtras `json:"extras"`
}

// AttributesExtras struct is the JSON format for misc asset attributes.
type AttributesExtras struct {
	State             string              `json:"status"`
	Company           string              `json:"company"`
	AssetType         string              `json:"assetType"` //chassis or server
	NetworkInterfaces *[]NetworkInterface `json:"network_interfaces"`
	LiveAssets        []string            `json:"live_assets"` //if its a chassis, we populate asset serials that are live.
}

// NetworkInterface struct is
type NetworkInterface struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address"`
}

// AssetFilter struct holds various attributes an asset should be filtered by.
//type AssetFilter struct {
//	Limit    int      //limit results
//	Offset   int      //query offset
//	Chassis  bool     //is a chassis asset
//	Server   bool     //is a server asset
//	Location []string //locations to return assets for
//	Company  []string //companys to return assets for
//	State    []string //filter assets by these states.
//}

func main() {

	serialsFlag := flag.NewFlagSet("serialsFlag", flag.ExitOnError)
	serialsArg := serialsFlag.String("serials", "", "--serials <foo>,<bar>")

	if len(os.Args) < 2 {
		log.Fatal("Usage: assetlookup [inventory --blah |enc --serials]")
	}

	switch os.Args[1] {
	case "enc":
		_ = serialsFlag.Parse(os.Args[2:])
		encCmd(*serialsArg)
	}
}

func encCmd(args string) {
	serials := strings.Split(args, ",")
	out, _ := json.Marshal(assetBySerial(serials))
	fmt.Print(string(out))
}

// Generally this would be a method that looks up the asset
// from the inventory using the serials.
// In this case we generate dummy assets and return AssetAttributes
func assetBySerial(serials []string) AssetAttributes {

	assets := make(map[string]Attributes)

	i := 1

	// generate dummy assets
	for _, serial := range serials {
		attributes := Attributes{
			Serial:    serial,
			Location:  "ams21",
			IPAddress: []string{fmt.Sprintf("192.168.0.%d", i)},
			Extras: AttributesExtras{
				State:     "live",
				Company:   "acme",
				AssetType: "server",
				NetworkInterfaces: &[]NetworkInterface{
					{
						Name:       "eth0",
						MACAddress: fmt.Sprintf("02:42:85:1f:80:0%d", i),
						IPAddress:  fmt.Sprintf("172.0.1.%d", i),
					},
				},
			},
		}

		i++
		assets[serial] = attributes
	}

	return AssetAttributes{Data: assets}

}
