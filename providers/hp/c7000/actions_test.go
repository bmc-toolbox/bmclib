package c7000

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

var (
	sshServer  net.Listener
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
		// "": []byte(``),
		// "": []byte(``),
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

func runSSHServer(config *ssh.ServerConfig, loading chan interface{}) {
	var err error
	sshServer, err = net.Listen("tcp", "127.0.0.1:2200")
	if err != nil {
		log.Fatalf("Failed to listen on 2200 (%s)", err)
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

func setupSSH() (bmc *C7000, err error) {
	setup()

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

	loading := make(chan interface{})
	go runSSHServer(config, loading)
	<-loading

	bmc, err = setup()
	if err != nil {
		return bmc, err
	}
	bmc.ip = "127.0.0.1:2200"

	return bmc, err
}

func tearDownSSH() {
	tearDown()
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
	expectedAnswer := 1

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.FindBladePosition("CZXXXXXXEK")
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

	answer, err := bmc.PowerCycleBlade(1)
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

	answer, err := bmc.ReseatBlade(1)
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

	answer, err := bmc.PowerOnBlade(1)
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

	answer, err := bmc.PowerOffBlade(1)
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

	answer, err := bmc.IsOnBlade(1)
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

	answer, err := bmc.PowerCycleBmcBlade(1)
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

	answer, err := bmc.PxeOnceBlade(1)
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

	answer, err := bmc.SetDynamicPower(false)
	if err != nil {
		t.Fatalf("Found errors calling bmc.SetDynamicPower %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
