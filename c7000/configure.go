package c7000

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ncode/bmc/cfgresources"
	"reflect"
)

func (c *C7000) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	cfg := reflect.ValueOf(config).Elem()

	//Each Field in ResourcesConfig struct is a ptr to a resource,
	//Here we figure the resources to be configured, i.e the ptr is not nil
	for r := 0; r < cfg.NumField(); r++ {
		resourceName := cfg.Type().Field(r).Name
		if cfg.Field(r).Pointer() != 0 {
			switch resourceName {
			case "User":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Syslog":
				syslogCfg := cfg.Field(r).Interface().(*cfgresources.Syslog)
				err := c.applySyslogParams(syslogCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       c.ip,
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ntp":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ldap":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ssl":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			default:
				log.WithFields(log.Fields{
					"step": "ApplyCfg",
				}).Warn("Unknown resource.")
				//fmt.Printf("%v\n", cfg.Field(r))

			}
		}
	}

	return err
}

// Applies syslog parameters
// theres no option to set the port
func (c *C7000) applySyslogParams(cfg *cfgresources.Syslog) (err error) {

	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Warn("Syslog resource expects parameter: Server.")
		return
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Debug("Syslog resource declared with enable: false.")
		return
	}

	//setup the XML payload
	server := Server{Text: cfg.Server}
	syslog := SetRemoteSyslogServer{Server: server}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(syslog, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Warn("Unable to marshal syslog payload.")
		return err
	}

	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Warn("Syslog apply request returned non 200.")
		return err
	}

	return err
}
