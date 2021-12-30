package ilo

import (
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/sshmock"
	"github.com/go-logr/logr"

	"testing"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

const (
	sshUsername = "super"
	sshPassword = "test"
)

var (
	sshAnswers = map[string][]byte{
		"power reset":    []byte(`Server resetting .......`),
		"reset /map1":    []byte(`Resetting iLO`),
		"power on":       []byte(`Server powering on .......`),
		"power off hard": []byte(`Forcing server power off .......`),
		"power":          []byte(`power: server power is currently: On`),
	}
)

func setupBMC() (func(), *Ilo, error) {
	ssh, err := sshmock.New(sshAnswers, logr.Discard())
	if err != nil {
		return nil, nil, err
	}
	tearDown, address, err := ssh.ListenAndServe()
	if err != nil {
		return nil, nil, err
	}

	sshClient, err := sshclient.New(address, sshUsername, sshPassword)
	if err != nil {
		return nil, nil, err
	}

	bmc := &Ilo{
		ip:        address,
		username:  sshUsername,
		password:  sshPassword,
		sshClient: sshClient,
	}

	return tearDown, bmc, err
}

func Test_ilo(t *testing.T) {
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
