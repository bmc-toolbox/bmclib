package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetBiosConfiguration(t *testing.T) {
	fixturePath := fixturesDir + "/v1/dell/bios.json"
	fh, err := os.Open(fixturePath)
	if err != nil {
		log.Fatalf("%s, failed to open fixture: %s", err.Error(), fixturePath)
	}

	defer fh.Close()

	b, err := io.ReadAll(fh)
	if err != nil {
		log.Fatalf("%s, failed to read fixture: %s", err.Error(), fixturePath)
	}

	var bios map[string]any
	err = json.Unmarshal([]byte(b), &bios)
	if err != nil {
		log.Fatalf("%s, failed to unmarshal fixture: %s", err.Error(), fixturePath)
	}

	expectedBiosConfig := make(map[string]string)
	for k, v := range bios["Attributes"].(map[string]any) {
		expectedBiosConfig[k] = fmt.Sprintf("%v", v)
	}

	tests := []struct {
		testName           string
		expectedBiosConfig map[string]string
	}{
		{
			"GetBiosConfiguration",
			expectedBiosConfig,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			biosConfig, err := mockClient.GetBiosConfiguration(context.TODO())
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedBiosConfig, biosConfig)
		})
	}
}
