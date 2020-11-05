package idrac9

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/sshmock"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
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

var _answers = map[string][]byte{
	"/sysmgmt/2015/bmc/info":    []byte(`{"Attributes":{"ADEnabled":"Disabled","BuildVersion":"37","FwVer":"3.15.15.15","GUITitleBar":"spare-H16Z4M2","IsOEMBranded":"0","License":"Enterprise","SSOEnabled":"Disabled","SecurityPolicyMessage":"By accessing this computer, you confirm that such access complies with your organization's security policy.","ServerGen":"14G","SrvPrcName":"NULL","SystemLockdown":"Disabled","SystemModelName":"PowerEdge M640","TFAEnabled":"Disabled","iDRACName":"spare-H16Z4M2"}}`),
	"/sysmgmt/2015/bmc/session": []byte(`{"status": "good", "authResult": 7, "forwardUrl": "something", "errorMsg": "none"}`),
}

func setupBMC() (func(), *IDrac9, error) {
	ssh, err := sshmock.New(sshAnswers)
	if err != nil {
		return nil, nil, err
	}
	tearDown, address, err := ssh.ListenAndServe()
	if err != nil {
		return nil, nil, err
	}
	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")

	for url := range _answers {
		url := url

		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(_answers[url])
		})
	}

	testLogger := logrus.New()
	bmc, err := New(context.TODO(), address, ip, sshUsername, sshPassword, logrusr.NewLogger(testLogger))
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
