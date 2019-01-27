package sshmock

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	mrand "math/rand"

	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

// New creates a new sshmock instance
func New(answers map[string][]byte, randomPort bool) (s *Server, err error) {
	mrand.Seed(time.Now().Unix())

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	port := 22
	if randomPort {
		port = mrand.Intn(4000-2000) + 2000
	}

	s = &Server{
		address: fmt.Sprintf("127.0.0.1:%d", port),
		wait:    make(chan interface{}),
		answers: answers,
	}

	key, err := s.generatePrivateKey(2048)
	if err != nil {
		return s, fmt.Errorf("failed to load private key: %s", err.Error())
	}

	private, err := ssh.ParsePrivateKey(s.encodePrivateKeyToPEM(key))
	if err != nil {
		return s, fmt.Errorf("failed to parse private key: %s", err.Error())
	}

	config.AddHostKey(private)

	go s.run(config)
	<-s.wait

	return s, err
}

// Server is the basic struct for the sshMock server
type Server struct {
	address string
	ssh     net.Listener
	answers map[string][]byte
	wait    chan interface{}
}

// Address returns the current sshmock server address
func (s *Server) Address() string {
	return s.address
}

func (s *Server) encodePrivateKeyToPEM(pk *rsa.PrivateKey) (payload []byte) {
	block := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(pk),
	}
	return pem.EncodeToMemory(&block)
}

func (s *Server) generatePrivateKey(bitSize int) (pk *rsa.PrivateKey, err error) {
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

func (s *Server) run(config *ssh.ServerConfig) {
	sshServer, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", s.address, err)
	}
	s.ssh = sshServer

	close(s.wait)
	for {
		conn, err := s.ssh.Accept()
		if err != nil {
			break
		}

		_, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		go ssh.DiscardRequests(reqs)
		go s.handleChannels(chans)
	}
}

func (s *Server) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

func (s *Server) handleChannel(newChannel ssh.NewChannel) {
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
				if answer, ok := s.answers[reqCmd.Text]; ok {
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
