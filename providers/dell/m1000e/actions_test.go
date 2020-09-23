package m1000e

import (
	"context"
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
		"racadm racreset": []byte(`CMC reset operation initiated successfully. It may take up to a minute 
			for the CMC to come back online again.
			`),
		"chassisaction powerup":   []byte(`Module power operation successful`),
		"chassisaction powerdown": []byte(`Module power operation successful`),
		"getsysinfo": []byte(`CMC Information:                         
			CMC Date/Time             = Tue Jan 04 2000 22:35      
			Primary CMC Location      = CMC-2      
			Primary CMC Version       = 6.10                                                                    
			Standby CMC Version       = 6.10                                                                           
			Last Firmware Update      = Mon Jan 03 2000 23:13
			Hardware Version          = A09                
											
			CMC Network Information:                                                                                  
			NIC Enabled               = 1       
			MAC Address               = 18:66:DA:9D:CD:CD        
			Register DNS CMC Name     = 0                                         
			DNS CMC Name              = cmc-5XXXXXX                                                                    
			Current DNS Domain        =             
			VLAN ID                   = 1
			VLAN Priority             = 0                                                                             
			VLAN Enabled              = 0         
																													
			CMC IPv4 Information:                                                                                      
			IPv4 Enabled              = 1                     
			Current IP Address        = 192.168.0.36
			Current IP Gateway        = 192.168.0.1                                                            
			Current IP Netmask        = 255.255.255.0                                                                  
			DHCP Enabled              = 1                                                                              
			Current DNS Server 1      = 0.0.0.0
			Current DNS Server 2      = 0.0.0.0                                                                        
			DNS Servers from DHCP     = 0     
																													
			CMC IPv6 Information:                                                                                     
			IPv6 Enabled              = 0                                                                             
			Autoconfiguration Enabled = 1                                                                             
			Link Local Address        = ::                                                                            
			Current IPv6 Address 1    = ::                                                                            
			Current IPv6 Gateway      = ::                   
			Current IPv6 DNS Server 1 = ::                                                                             
			Current IPv6 DNS Server 2 = ::                           
			DNS Servers from DHCPv6   = 1                                                                             
																				
			Chassis Information:                                 
			System Model              = PowerEdge M1000e
			System AssetTag           = 00000                                                                          
			Service Tag               = 5XXXXXX      
			Chassis Name              = CMC-5XXXXXX                
			Chassis Location          = [UNDEFINED]
			Chassis Midplane Version  = 1.1                                                                     
			Power Status              = ON                                                                             
			System ID                 = 1486                 
	        `),
		"getsvctag": []byte(`<Module>        <ServiceTag>
			Chassis         5XXXXXX
			Switch-1        0000000
			Switch-2        N/A
			Switch-3        N/A
			Switch-4        N/A
			Switch-5        N/A
			Switch-6        N/A
			Server-1        N/A
			Server-2        74XXX72
			Server-3        N/A
			Server-4        N/A
			Server-5        N/A
			Server-6        N/A
			Server-7        N/A
			Server-8        N/A
			Server-9        N/A
			Server-10       N/A
			Server-11       N/A
			Server-12       N/A
			Server-13       N/A
			Server-14       N/A
			Server-15       N/A
			Server-16       N/A
			`),
		"serveraction -m server-2 hardreset":   []byte(`Server power operation successful`),
		"serveraction -m server-2 reseat -f":   []byte(`Server power operation successful`),
		"serveraction -m server-2 powerup":     []byte(`Server power operation successful`),
		"serveraction -m server-2 powerdown":   []byte(`Server power operation successful`),
		"serveraction -m server-2 powerstatus": []byte(`ON`),
		"serveraction -m server-1 powerstatus": []byte(`OFF`),
		"racreset -m server-2": []byte(`RAC reset operation initiated successfully for server-2.
			It may take up to a minute for the RAC(s) to come back online again.`),
		"deploy -m server-2 -b PXE -o yes":                                    []byte(`The blade was deployed successfully.`),
		"config -g cfgServerInfo -o cfgServerIPMIOverLanEnable -i 2 1":        []byte(`Object value modified successfully`),
		"config -g cfgChassisPower -o cfgChassisDynamicPSUEngagementEnable 1": []byte(`Object value modified successfully`),
		"racadm setflexaddr -i 1 0": []byte(`Slot 2 FlexAddress state set successfully.
			This will force a reset on hardware affected by the Flex Address change.
			Please wait for up to a few minutes before performing additional power
			related actions (eg. reset, powerdown) on the affected hardware.
			`),
	}
)

func setupBMC() (func(), *M1000e, error) {
	ssh, err := sshmock.New(sshAnswers)
	if err != nil {
		return nil, nil, err
	}
	tearDown, address, err := ssh.ListenAndServe()
	if err != nil {
		return nil, nil, err
	}

	testLogger := logrus.New()
	bmc, err := New(context.TODO(), address, sshUsername, sshPassword, logrusr.NewLogger(testLogger))
	if err != nil {
		tearDown()
		return nil, nil, err
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

func Test_FindBladePosition(t *testing.T) {
	tearDown, bmc, err := setupBMC()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDown()

	want, wantErr := 2, false

	got, err := bmc.FindBladePosition("74XXX72")

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
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

	got, err := bmc.PowerCycleBlade(2)

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

	got, err := bmc.ReseatBlade(2)

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

	got, err := bmc.PowerOnBlade(2)

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

	got, err := bmc.PowerOffBlade(2)

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

	got, err := bmc.IsOnBlade(2)

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

	got, err := bmc.PowerCycleBmcBlade(2)

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

	got, err := bmc.PxeOnceBlade(2)

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

	want, wantErr := true, false

	got, err := bmc.SetIpmiOverLan(2, true)

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

	got, err := bmc.SetDynamicPower(true)

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

	want, wantErr := true, false

	got, err := bmc.SetFlexAddressState(1, false)

	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}

	if got != want {
		t.Errorf("got = %v, want %v", got, want)
	}
}
