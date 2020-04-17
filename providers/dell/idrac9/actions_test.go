package idrac9

import (
	"testing"

	"github.com/bmc-toolbox/bmclib/sshmock"
)

const (
	sshUsername = "super"
	sshPassword = "test"
)

var (
	sshAnswers = map[string][]byte{
		"racadm serveraction hardreset": []byte(`Server power operation successful`),
		"racadm racreset hard": []byte(`RAC reset operation initiated successfully. It may take a few
			minutes for the RAC to come online again.
		   `),
		"racadm serveraction powerup":     []byte(`Server power operation successful`),
		"racadm serveraction powerdown":   []byte(`Server power operation successful`),
		"racadm serveraction powerstatus": []byte(`Server power status: ON`),
		"racadm config -g cfgServerInfo -o cfgServerBootOnce 1": []byte(`Object value modified successfully


			RAC1169: The RACADM "config" command will be deprecated in a
			future version of iDRAC firmware. Run the RACADM 
			"racadm set" command to configure the iDRAC configuration parameters.
			For more information on the set command, run the RACADM command
			"racadm help set".
			
			`),
		"racadm config -g cfgServerInfo -o cfgServerFirstBootDevice PXE": []byte(`Object value modified successfully


			RAC1169: The RACADM "config" command will be deprecated in a
			future version of iDRAC firmware. Run the RACADM 
			"racadm set" command to configure the iDRAC configuration parameters.
			For more information on the set command, run the RACADM command
			"racadm help set".
			
			`),
	}
)

func setupBMC() (func(), *IDrac9, error) {
	ssh, err := sshmock.New(sshAnswers)
	if err != nil {
		return nil, nil, err
	}
	tearDown, address, err := ssh.ListenAndServe()
	if err != nil {
		return nil, nil, err
	}

	bmc, err := New(address, sshUsername, sshPassword)
	if err != nil {
		tearDown()
		return nil, nil, err
	}

	return tearDown, bmc, err
}

func Test_IDrac9(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("failed to setup BMC: %v", err)
	}
	defer tearDown()

	tests := []struct {
		name      string
		bmcMethod func() (bool, error)
		want      bool
		wantErr   bool
	}{
		{
			name:      "PowerCycle",
			bmcMethod: bmc.PowerCycle,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "PowerCycleBmc",
			bmcMethod: bmc.PowerCycleBmc,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "PowerOn",
			bmcMethod: bmc.PowerOn,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "PowerOff",
			bmcMethod: bmc.PowerOff,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "PxeOnce",
			bmcMethod: bmc.PxeOnce,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "IsOn",
			bmcMethod: bmc.IsOn,
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.bmcMethod()

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
