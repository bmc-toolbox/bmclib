package inventory

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"
)

// Enc struct holds attributes required to run inventory/enc methods.
type Enc struct {
	Log             *logrus.Logger
	BatchSize       int
	AssetsChan      chan<- []asset.Asset
	MetricsEmitter  *metrics.Emitter
	Config          *config.Params
	FilterAssetType []string
	StopChan        <-chan struct{}
}

// AssetAttributes is used to unmarshal data returned from an ENC.
type AssetAttributes struct {
	Data        map[string]Attributes `json:"data"` //map of asset IPs/Serials to attributes
	EndOfAssets bool                  `json:"end_of_assets"`
}

// Attributes is used to unmarshal data returned from an ENC.
type Attributes struct {
	Location          string              `json:"location"`
	NetworkInterfaces *[]NetworkInterface `json:"network_interfaces"`
	BMCIPAddress      []string            `json:"-"`
	Extras            *AttributesExtras   `json:"extras"`
}

// NetworkInterface is used to unmarshal data returned from the ENC.
type NetworkInterface struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac_address"`
	IPAddress  string `json:"ip_address"`
}

// AttributesExtras is used to unmarshal data returned from an ENC.
type AttributesExtras struct {
	State   string `json:"status"`
	Company string `json:"company"`
	//if its a chassis, this would hold serials for blades in the live state
	LiveAssets *[]string `json:"live_assets,omitempty"`
}

func stringHasPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(s), prefix) {
			return true
		}
	}

	return false
}

// SetBMCInterfaces populates IPAddresses of BMC interfaces,
// from the Slice of NetworkInterfaces
func (e *Enc) SetBMCInterfaces(attributes Attributes) Attributes {

	if attributes.NetworkInterfaces == nil {
		return attributes
	}

	bmcNicPrefixes := e.Config.InventoryParams.BMCNicPrefix
	for _, nic := range *attributes.NetworkInterfaces {
		if stringHasPrefix(nic.Name, bmcNicPrefixes) && nic.IPAddress != "" {
			attributes.BMCIPAddress = append(attributes.BMCIPAddress, nic.IPAddress)
		}
	}

	return attributes
}

// AttributesExtrasAsMap accepts a AttributesExtras struct as input,
// and returns all attributes as a map
func AttributesExtrasAsMap(attributeExtras *AttributesExtras) (extras map[string]string) {

	extras = make(map[string]string)

	extras["state"] = strings.ToLower(attributeExtras.State)
	extras["company"] = strings.ToLower(attributeExtras.Company)

	if attributeExtras.LiveAssets != nil {
		extras["liveAssets"] = strings.ToLower(strings.Join(*attributeExtras.LiveAssets, ","))
	} else {
		extras["liveAssets"] = ""
	}

	return extras
}

//AssetRetrieve looks at c.Config.FilterParams
//and returns the appropriate function that will retrieve assets.
func (e *Enc) AssetRetrieve() func() {

	//setup the asset types we want to retrieve data for.
	switch {
	case e.Config.FilterParams.Chassis:
		e.FilterAssetType = append(e.FilterAssetType, "chassis")
	case e.Config.FilterParams.Servers:
		e.FilterAssetType = append(e.FilterAssetType, "servers")
	case !e.Config.FilterParams.Chassis && !e.Config.FilterParams.Servers:
		e.FilterAssetType = []string{"chassis", "servers"}
	}

	//Based on the filter param given, return the asset iterator method.
	switch {
	case e.Config.FilterParams.Serials != "":
		return e.AssetIterBySerial
	case e.Config.FilterParams.Ips != "":
		return e.AssetIterByIP
	default:
		return e.AssetIter
	}
}

// ExecCmd executes the executable with the given args and returns
// the response as a slice of bytes, and the error if any.
func ExecCmd(exe string, args []string) (out []byte, err error) {

	cmd := exec.Command(exe, args...)

	//To ignore SIGINTs received by bmcbutler,
	//the commands are spawned in its own process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	out, err = cmd.Output()
	if err != nil {
		return out, err
	}

	return out, err
}

// SetChassisInstalled is a method used to update a chassis state in the inventory.
func (e *Enc) SetChassisInstalled(serials string) {

	log := e.Log
	component := "SetChassisInstalled"

	//assetlookup inventory --set-chassis-installed FOO123,BAR123
	cmdArgs := []string{"inventory", "--set-chassis-installed", serials}

	encBin := e.Config.InventoryParams.EncExecutable
	out, err := ExecCmd(encBin, cmdArgs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Warn("Command to update chassis state returned error.")
	}
}

// nolint: gocyclo
func (e *Enc) encQueryBySerial(serials string) (assets []asset.Asset) {

	log := e.Log
	metric := e.MetricsEmitter
	component := "encQueryBySerial"

	//assetlookup enc --serials FOO123,BAR123
	cmdArgs := []string{"enc", "--serials", serials}

	encBin := e.Config.InventoryParams.EncExecutable
	out, err := ExecCmd(encBin, cmdArgs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Fatal("Inventory query failed, lookup command returned error.")
	}

	cmdResp := AssetAttributes{}
	err = json.Unmarshal(out, &cmdResp)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Fatal("JSON Unmarshal command response returned error.")
	}

	if len(cmdResp.Data) == 0 {
		log.WithFields(logrus.Fields{
			"component": component,
			"Serial(s)": serials,
		}).Warn("No assets returned by inventory for given serial(s).")

		return []asset.Asset{}
	}

	missingSerials := strings.Split(serials, ",")
	for serial, attributes := range cmdResp.Data {

		attributes := e.SetBMCInterfaces(attributes)
		if len(attributes.BMCIPAddress) == 0 {
			metric.IncrCounter([]string{"inventory", "assets_noip_enc"}, 1)
			continue
		}

		// missing Serials are Serials we looked up using the enc and got no data for.
		for idx, s := range missingSerials {
			if s == serial {
				// if its in the list, purge it.
				missingSerials = append(missingSerials[:idx], missingSerials[idx+1:]...)
			}
		}

		extras := AttributesExtrasAsMap(attributes.Extras)
		assets = append(assets,
			asset.Asset{IPAddresses: attributes.BMCIPAddress,
				Serial:   serial,
				Location: attributes.Location,
				Extra:    extras,
			})
	}

	// append missing Serials to assets
	if len(missingSerials) > 0 {
		for _, serial := range missingSerials {
			assets = append(assets, asset.Asset{Serial: serial, IPAddresses: []string{}})
		}
	}

	if metric != nil {
		metric.IncrCounter([]string{"inventory", "assets_fetched_enc"}, float32(len(assets)))
	}

	return assets
}

// nolint: gocyclo
func (e *Enc) encQueryByIP(ips string) (assets []asset.Asset) {

	log := e.Log
	metric := e.MetricsEmitter
	component := "encQueryByIP"

	// if no attributes can be received we return assets objs
	// populate and return slice of assets with no attributes except ips.
	populateAssetsWithNoAttributes := func() {
		ipList := strings.Split(ips, ",")
		for _, ip := range ipList {
			assets = append(assets, asset.Asset{IPAddresses: []string{ip}})
		}
	}

	//assetlookup enc --serials 192.168.1.1,192.168.1.2
	cmdArgs := []string{"enc", "--ips", ips}

	encBin := e.Config.InventoryParams.EncExecutable
	out, err := ExecCmd(encBin, cmdArgs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Warn("Inventory query failed, lookup command returned error.")

		populateAssetsWithNoAttributes()
		return assets
	}

	cmdResp := AssetAttributes{}
	err = json.Unmarshal(out, &cmdResp)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Fatal("JSON Unmarshal command response returned error.")
	}

	if len(cmdResp.Data) == 0 {
		log.WithFields(logrus.Fields{
			"component": component,
			"IP(s)":     ips,
		}).Debug("No assets returned by inventory for given IP(s).")

		populateAssetsWithNoAttributes()
		return assets
	}

	// missing IPs are IPs we looked up using the enc and got no data for.
	missingIPs := strings.Split(ips, ",")
	for serial, attributes := range cmdResp.Data {

		attributes := e.SetBMCInterfaces(attributes)
		if len(attributes.BMCIPAddress) == 0 {
			populateAssetsWithNoAttributes()
			metric.IncrCounter([]string{"inventory", "assets_noip_enc"}, 1)
			continue
		}

		for _, bmcIPAddress := range attributes.BMCIPAddress {
			for idx, ip := range missingIPs {
				if ip == bmcIPAddress {
					missingIPs = append(missingIPs[:idx], missingIPs[idx+1:]...)
				}
			}
		}
		extras := AttributesExtrasAsMap(attributes.Extras)

		assets = append(assets,
			asset.Asset{IPAddresses: attributes.BMCIPAddress,
				Serial:   serial,
				Location: attributes.Location,
				Extra:    extras,
			})
	}

	// append missing IPs.
	if len(missingIPs) > 0 {
		for _, ip := range missingIPs {
			assets = append(assets, asset.Asset{IPAddresses: []string{ip}})
		}
	}

	metric.IncrCounter([]string{"inventory", "assets_fetched_enc"}, float32(len(assets)))

	return assets
}

// encQueryByOffset returns a slice of assets and if the query reached the end of assets.
// assetType is one of 'servers/chassis'
// location is a comma delimited list of locations
func (e *Enc) encQueryByOffset(assetType string, offset int, limit int, location string) (assets []asset.Asset, endOfAssets bool) {

	component := "EncQueryByOffset"
	metric := e.MetricsEmitter
	log := e.Log

	assets = make([]asset.Asset, 0)

	var encAssetTypeFlag string

	switch assetType {
	case "servers":
		encAssetTypeFlag = "--server"
	case "chassis":
		encAssetTypeFlag = "--chassis"
	case "discretes":
		encAssetTypeFlag = "--server"
	}

	//assetlookup inventory --server --offset 0 --limit 10
	cmdArgs := []string{"inventory", encAssetTypeFlag,
		"--limit", strconv.Itoa(limit),
		"--offset", strconv.Itoa(offset)}

	//--location ams9
	if location != "" {
		cmdArgs = append(cmdArgs, "--location")
		cmdArgs = append(cmdArgs, location)
	}

	encBin := e.Config.InventoryParams.EncExecutable
	out, err := ExecCmd(encBin, cmdArgs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Fatal("Inventory query failed, lookup command returned error.")
	}

	cmdResp := AssetAttributes{}
	err = json.Unmarshal(out, &cmdResp)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
			"cmd":       fmt.Sprintf("%s %s", encBin, strings.Join(cmdArgs, " ")),
			"output":    fmt.Sprintf("%s", out),
		}).Fatal("JSON Unmarshal command response returned error.")
	}

	endOfAssets = cmdResp.EndOfAssets

	if len(cmdResp.Data) == 0 {
		return []asset.Asset{}, endOfAssets
	}

	for serial, attributes := range cmdResp.Data {

		attributes := e.SetBMCInterfaces(attributes)
		if len(attributes.BMCIPAddress) == 0 {
			metric.IncrCounter([]string{"inventory", "assets_noip_enc"}, 1)
			continue
		}

		extras := AttributesExtrasAsMap(attributes.Extras)
		assets = append(assets,
			asset.Asset{IPAddresses: attributes.BMCIPAddress,
				Serial:   serial,
				Type:     assetType,
				Location: attributes.Location,
				Extra:    extras,
			})
	}

	metric.IncrCounter([]string{"inventory", "assets_fetched_enc"}, float32(len(assets)))

	return assets, endOfAssets
}

// AssetIter fetches assets and sends them over the asset channel.
func (e *Enc) AssetIter() {

	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel
	//component := "AssetIterEnc"

	//metric := d.MetricsEmitter

	var interrupt bool

	go func() { <-e.StopChan; interrupt = true }()

	defer close(e.AssetsChan)
	//defer d.MetricsEmitter.MeasureSince(component, time.Now())

	locations := strings.Join(e.Config.Locations, ",")
	for _, assetType := range e.FilterAssetType {

		var limit = e.BatchSize
		var offset = 0

		for {
			var endOfAssets bool

			assets, endOfAssets := e.encQueryByOffset(assetType, offset, limit, locations)

			e.Log.WithFields(logrus.Fields{
				"component": "inventory",
				"method":    "AssetIter",
				"Asset":     assetType,
				"Offset":    offset,
				"Limit":     limit,
				"locations": locations,
			}).Debug("Assets retrieved.")

			//pass the asset to the channel
			e.AssetsChan <- assets

			//increment offset for next set of assets
			offset += limit

			//If the ENC indicates we've reached the end of assets
			if endOfAssets || interrupt {

				e.Log.WithFields(logrus.Fields{
					"component": "inventory",
					"method":    "AssetIter",
				}).Debug("Reached end of assets/interrupt received.")
				break
			}
		} // endless for
	} // for each assetType
}

// AssetIterBySerial reads in list of serials passed in via cli,
// queries the ENC for the serials, passes them to the assets channel
func (e *Enc) AssetIterBySerial() {

	defer close(e.AssetsChan)

	//get serials passed in via cli - they need to be comma separated
	serials := e.Config.FilterParams.Serials

	//query ENC for given serials
	assets := e.encQueryBySerial(serials)

	//pass assets returned by ENC to the assets channel
	e.AssetsChan <- assets
}

// AssetIterByIP reads in list of ips passed in via cli,
// queries the ENC for attributes related to the, passes them to the assets channel
// if no attributes for a given IP are returned, an asset with just the IP is returned.
func (e *Enc) AssetIterByIP() {

	defer close(e.AssetsChan)

	//get ips passed in via cli - they need to be comma separated
	ips := e.Config.FilterParams.Ips

	//query ENC for given serials
	assets := e.encQueryByIP(ips)

	//pass assets returned by ENC to the assets channel
	e.AssetsChan <- assets
}
