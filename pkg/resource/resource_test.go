package resource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/sirupsen/logrus"
)

// Test ReadYamlTemplate method loads expected yaml data.
func TestReadYamlTemplate(t *testing.T) {

	resourceConfig := "../../samples/cfg/configuration.yml"

	// validate ReadYamlTemplate returns byte slice
	configBytes, err := ReadYamlTemplate(resourceConfig)
	if err != nil {
		t.Fatalf("Error reading config file %s : %s", resourceConfig, err)
	}

	if len(configBytes) < 1 {
		t.Fatal("Resource config returned empty byte slice")
	}

	if !strings.Contains(string(configBytes), "name: Administrator") {
		t.Fatal("Expected string not found in template.")
	}
}

// Test template rendering
func TestRenderYamlTemplate(t *testing.T) {

	r := Resource{
		Log: logrus.New(),
		Asset: &asset.Asset{
			Serial: "FOOBAR",
			Vendor: "ACME",
			Model:  "002",
			Type:   "Server",
		},
	}

	resourceConfig := "../../samples/cfg/configuration.yml"

	// read in resource config template
	configBytes, err := ReadYamlTemplate(resourceConfig)
	if err != nil {
		t.Fatalf("Error reading config file %s : %s", resourceConfig, err)
	}

	// render as plush template
	rendered := r.RenderYamlTemplate(configBytes)
	if !strings.Contains(string(rendered), "cn=acme,cn=bmcUsers") {
		t.Fatal("Expected string not found in rendered template")
	}

}

// Test yaml unmarshal and LoadConfigResources returns a cfgresources.ResourcesConfig instance
func TestLoadConfigResources(t *testing.T) {

	resourceConfig := "../../samples/cfg/configuration.yml"

	// read in resource config template
	configBytes, err := ReadYamlTemplate(resourceConfig)
	if err != nil {
		t.Fatalf("Error reading config file %s : %s", resourceConfig, err)
	}

	r := Resource{
		Log: logrus.New(),
		Asset: &asset.Asset{
			Serial: "FOOBAR",
			Vendor: "ACME",
			Model:  "002",
			Type:   "Server",
		},
	}

	configResources := r.LoadConfigResources(configBytes)
	if fmt.Sprintf("%T", configResources) != "*cfgresources.ResourcesConfig" {
		t.Fatal("Expected return type does not match *cfgresources.ResourcesConfig")
	}

	if configResources.LdapGroup[0].Group != "cn=acme,cn=bmcAdmins" {
		t.Fatal("Expected string not found in LdapGroup config resource")
	}

}
