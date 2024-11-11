package dell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvFirmwareTaskOem(t *testing.T) {
	testCases := []struct {
		name        string
		oemdata     []byte
		expectedJob oem
		expectedErr string
	}{
		{
			name: "Valid OEM data",
			oemdata: []byte(`{
				"Dell": {
					"@odata.type": "#DellJob.v1_4_0.DellJob",
					"CompletionTime": null,
					"Description": "Job Instance",
					"EndTime": "TIME_NA",
					"Id": "JID_005950769310",
					"JobState": "Scheduled",
					"JobType": "FirmwareUpdate",
					"Message": "Task successfully scheduled.",
					"MessageArgs": [],
					"MessageId": "IDRAC.2.8.JCP001",
					"Name": "Firmware Update: BIOS",
					"PercentComplete": 0,
					"StartTime": "TIME_NOW",
					"TargetSettingsURI": null
				}
			}`),
			expectedJob: oem{
				Dell{
					OdataType:         "#DellJob.v1_4_0.DellJob",
					CompletionTime:    nil,
					Description:       "Job Instance",
					EndTime:           "TIME_NA",
					ID:                "JID_005950769310",
					JobState:          "Scheduled",
					JobType:           "FirmwareUpdate",
					Message:           "Task successfully scheduled.",
					MessageArgs:       []interface{}{},
					MessageID:         "IDRAC.2.8.JCP001",
					Name:              "Firmware Update: BIOS",
					PercentComplete:   0,
					StartTime:         "TIME_NOW",
					TargetSettingsURI: nil,
				},
			},
			expectedErr: "",
		},
		{
			name:        "Empty OEM data",
			oemdata:     []byte(`{}`),
			expectedJob: oem{},
			expectedErr: "empty oem data",
		},
		{
			name:        "Invalid OEM data",
			oemdata:     []byte(`{"InvalidKey": "InvalidValue"}`),
			expectedJob: oem{},
			expectedErr: "invalid oem data",
		},
		{
			name: "Unexpected job type",
			oemdata: []byte(`{
					"Dell": {
						"JobType": "InvalidJobType",
						"Description": "Job Instance",
						"JobState": "Scheduled"
					}
				}`),
			expectedJob: oem{},
			expectedErr: "unexpected job type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			job, err := convFirmwareTaskOem(tc.oemdata)
			if tc.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedJob, job)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}
		})
	}
}
