package m1000e

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	mrand "math/rand"

	"fmt"
	"log"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

func init() {
	mrand.Seed(time.Now().Unix())
}

func sshServerAddress(min, max int) string {
	return fmt.Sprintf("127.0.0.1:%d", mrand.Intn(max-min)+min)
}

var (
	sshServer  net.Listener
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

func generatePrivateKey(bitSize int) (pk *rsa.PrivateKey, err error) {
	pk, err = rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return pk, err
	}

	err = pk.Validate()
	if err != nil {
		return pk, err
	}

	return pk, err
}

func encodePrivateKeyToPEM(pk *rsa.PrivateKey) (payload []byte) {
	block := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(pk),
	}
	return pem.EncodeToMemory(&block)
}

func runSSHServer(config *ssh.ServerConfig, address string, loading chan interface{}) {
	var err error
	sshServer, err = net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", address, err)
	}

	close(loading)
	for {
		conn, err := sshServer.Accept()
		if err != nil {
			break
		}

		_, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "exec"
	// We just want to handle "exec".
	go func() {
		for req := range requests {
			switch req.Type {
			case "exec":
				var reqCmd struct{ Text string }
				if err := ssh.Unmarshal(req.Payload, &reqCmd); err != nil {
					log.Printf("failed: %v\n", err)
				}
				if answer, ok := sshAnswers[reqCmd.Text]; ok {
					if len(answer) == 0 {
						channel.Stderr().Write([]byte(fmt.Sprintf("answer empty for %s", reqCmd.Text)))
						req.Reply(req.WantReply, nil)
						if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1}); err != nil {
							log.Printf("failed: %v\n", err)
						}
					} else {
						channel.Write(answer)
						req.Reply(req.WantReply, nil)
						if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
							log.Printf("failed: %v\n", err)
						}
					}
				} else {
					channel.Stderr().Write([]byte(fmt.Sprintf("answer not found for %s", reqCmd.Text)))
					req.Reply(req.WantReply, nil)
					if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1}); err != nil {
						log.Printf("failed: %v\n", err)
					}
				}
				if err := channel.Close(); err != nil {
					log.Printf("failed: %v\n", err)
				}
			default:
				fmt.Println(req.Type)
			}
		}
	}()
}

func setupSSH() (bmc *M1000e, err error) {
	username := "super"
	password := "test"

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	key, err := generatePrivateKey(2048)
	if err != nil {
		log.Fatal("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(encodePrivateKeyToPEM(key))
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	address := sshServerAddress(2000, 4000)
	loading := make(chan interface{})
	go runSSHServer(config, address, loading)
	<-loading

	bmc, err = New(address, username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDownSSH() {
	sshServer.Close()
}

func TestChassisPowerCycle(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerCycle()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycle %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerOff(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOff()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOff %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisIsOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.IsOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
func TestChassisFindBladePosition(t *testing.T) {
	expectedAnswer := 2

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.FindBladePosition("74XXX72")
	if err != nil {
		t.Fatalf("Found errors calling bmc.FindBladePosition %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerCycleBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerCycleBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycleBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisReseatBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.ReseatBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.ReseatBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerOnBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOnBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOnBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerOffBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOffBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOffBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisIsOnBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.IsOnBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsOnBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPowerCycleBmcBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerCycleBmcBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycleBmcBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisPxeOnceBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PxeOnceBlade(2)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PxeOnceBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisSetIpmiOverLan(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.SetIpmiOverLan(2, true)
	if err != nil {
		t.Fatalf("Found errors calling bmc.PxeOnceBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisSetDynamicPower(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.SetDynamicPower(true)
	if err != nil {
		t.Fatalf("Found errors calling bmc.SetDynamicPower %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestChassisSetFlexAddressState(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.SetFlexAddressState(1, false)
	if err != nil {
		t.Fatalf("Found errors calling bmc.SetFlexAddressState %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
