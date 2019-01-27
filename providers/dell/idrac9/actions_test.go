package idrac9

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

func setupSSH() (bmc *IDrac9, err error) {
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

func TestIDracPowerCycle(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerCycle()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycle %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerCycleBmc(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerCycleBmc()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycleBmc %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerOff(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerOff()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOff %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPxeOnce(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PxeOnce()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PxeOnce %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracIsOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.IsOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}
