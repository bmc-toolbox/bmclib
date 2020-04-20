package c7000

import (
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/sshmock"

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
		"RESTART OA ACTIVE": []byte(`Restarting Onboard Administrator in bay`),
		"SHOW SERVER NAMES": []byte(`
			Bay Server Name                                       Serial Number   Status   Power   UID Partner
			--- ------------------------------------------------- --------------- -------- ------- --- -------
			  1 fdi                                               CZXXXXXXEK      OK       On      Off 
			  2 [Absent]                                       
			  3 [Absent]                                        
			  4 [Absent]                                        
			  5 [Absent]                                         
			  6 [Absent]                                          
			  7 [Absent]                                          
			  8 [Absent]                                          
			  9 [Absent]                                          
			 10 [Absent]                                          
			 11 [Absent]                                          
			 12 [Absent]                                          
			 13 [Absent]                                          
			 14 [Absent]                                          
			 15 [Absent]                                          
			 16 [Absent]                                          
			Totals: 1 server blades installed, 1 powered on.
			`),
		"REBOOT SERVER 1 FORCE":   []byte(`Forcing reboot of Blade 1`),
		"RESET SERVER 1":          []byte(`Successfully reset the E-Fuse for device bay 1.`),
		"POWERON SERVER 1":        []byte(`Powering on blade 1.`),
		"POWEROFF SERVER 1 FORCE": []byte(`Blade 1 is powering down.`),
		"SHOW SERVER STATUS 1": []byte(`Blade #1 Status:
			Power: On
			Current Wattage used: 500
			Health: OK
			Unit Identification LED: Off
			Virtual Fan: 0%
			Diagnostic Status:
					Internal Data                            OK
					Management Processor                     OK
					I/O Configuration                        OK
					Power                                    OK
					Cooling                                  OK
					Device Failure                           OK
					Device Degraded                          OK
					iLO Network                              OK
					Mezzanine Card                           OK
        	`),
		"RESET ILO 1":                []byte(`Bay 1: Successfully reset iLO through Hardware reset`),
		"SET SERVER BOOT ONCE PXE 1": []byte(`Blade #1 boot order changed to PXE`),
		"SET POWER SAVINGS OFF": []byte(`Power Settings were updated to:

			Power Mode: Redundant
			Dynamic Power: Disabled
			Set Power Limit: Not Set
			
			Power Capacity:              5300 Watts DC
			Power Available:             2767 Watts DC
			Power Allocated:             2533 Watts DC
			Present Power:                884 Watts AC
			Power Limit:                 6445 Watts AC			
			`),
		"SHOW OA INFO": []byte(`Onboard Administrator #1 information:
			Product Name  : BladeSystem c7000 DDR2 Onboard Administrator with KVM
			Part Number   : XXXXXX-XXX
			Spare Part No.: XXXXXX-XXX
			Serial Number : OXXXXXXX74    
			UUID          : 09OB2XXXXXXX4    
			Manufacturer  : HP
			Firmware Ver. : 4.80 Dec 13 2017
			Hw Board Type : 2
			Hw Version    : A1
			Loader Version: U-Boot 1.2.0 (Aug 24 2011 - 14:22:07)
			Serial Port:
				Baud Rate   : 9600
				Parity      : None
				Data bits   : 8
				Stop bits   : 1
				Flow control: None
	
	`),
	}
)

func setupBMC() (func(), *C7000, error) {
	ssh, err := sshmock.New(sshAnswers)
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

	bmc := &C7000{
		ip:        address,
		username:  sshUsername,
		password:  sshPassword,
		sshClient: sshClient,
	}

	return tearDown, bmc, err
}

func Test_chassisBMC(t *testing.T) {
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
			name:      "PowerOn",
			bmcMethod: bmc.PowerOn,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "PowerOff",
			bmcMethod: bmc.PowerOff,
			want:      false,
			wantErr:   true,
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

func Test_FindBladePosition(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := 1, false

	got, err := bmc.FindBladePosition("CZXXXXXXEK")

	if err != nil {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if (err != nil) != wantErr {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_PowerCycleBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.PowerCycleBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_ReseatBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.ReseatBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_PowerOnBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.PowerOnBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_PowerOffBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.PowerOffBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_IsOnBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.IsOnBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_PowerCycleBmcBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.PowerCycleBmcBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_PxeOnceBlade(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.PxeOnceBlade(1)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_SetIpmiOverLan(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := false, true

	got, err := bmc.SetIpmiOverLan(1, true)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_SetDynamicPower(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := true, false

	got, err := bmc.SetDynamicPower(false)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}

func Test_SetFlexAddressState(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := false, true

	got, err := bmc.SetFlexAddressState(1, false)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}
