package idrac8

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

var (
	sshServer  net.Listener
	sshAnswers = map[string][]byte{
		"racadm serveraction hardreset": []byte(`Server power operation successful\n`),
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

func runSSHServer(config *ssh.ServerConfig) {
	var err error
	sshServer, err = net.Listen("tcp", "127.0.0.1:2200")
	if err != nil {
		log.Fatalf("Failed to listen on 2200 (%s)", err)
	}

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

func setupSSH() (bmc *IDrac8, err error) {
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

	go runSSHServer(config)
	time.Sleep(1 * time.Second)

	bmc, err = New("127.0.0.1:2200", username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
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
				channel.Write(sshAnswers[reqCmd.Text])
				if req.WantReply {
					req.Reply(true, nil)
				}
				if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
					log.Printf("failed: %v\n", err)
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
