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

const (
	privateKeyBitSize           = 2048
	maxAttemptsToCreateListener = 10
)

// Server is the basic struct for the sshMock server
type Server struct {
	config  *ssh.ServerConfig
	answers map[string][]byte
}

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

// New creates a new sshmock instance
func New(answers map[string][]byte) (*Server, error) {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	privateKey, err := generateHostKey()
	if err != nil {
		return nil, err
	}

	config.AddHostKey(privateKey)

	server := &Server{
		config:  config,
		answers: answers,
	}

	return server, err
}

func (s *Server) ListenAndServe() (func(), string, error) {
	listener, err := createListener()
	if err != nil {
		return nil, "", err
	}

	addr := listener.Addr().String()
	shutdown := func() { listener.Close() }

	go s.run(listener)

	return shutdown, addr, nil
}

func (s *Server) run(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}

		if err := s.handleConnection(conn); err != nil {
			log.Printf("failed to handle connection: %v", err)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) error {
	defer conn.Close()
	serverConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		return err
	}
	defer serverConn.Close()

	go ssh.DiscardRequests(reqs)
	s.handleChannels(chans)

	return nil
}

func (s *Server) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go s.handleChannel(newChannel)
	}
}

func (s *Server) handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	// Sessions have out-of-band requests such as "shell", "pty-req" and "exec"
	// We just want to handle "exec".
	for req := range requests {
		if req.Type != "exec" {
			continue
		}
		var reqCmd struct{ Text string }
		if err := ssh.Unmarshal(req.Payload, &reqCmd); err != nil {
			log.Printf("failed: %v\n", err)
		}
		if answer, ok := s.answers[reqCmd.Text]; ok {
			if len(answer) == 0 {
				_, _ = channel.Stderr().Write([]byte(fmt.Sprintf("answer empty for %s", reqCmd.Text)))
				_ = req.Reply(req.WantReply, nil)
				if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1}); err != nil {
					log.Printf("failed: %v\n", err)
				}
			} else {
				_, _ = channel.Write(answer)
				_ = req.Reply(req.WantReply, nil)
				if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
					log.Printf("failed: %v\n", err)
				}
			}
		} else {
			_, _ = channel.Stderr().Write([]byte(fmt.Sprintf("answer not found for %s", reqCmd.Text)))
			_ = req.Reply(req.WantReply, nil)
			if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1}); err != nil {
				log.Printf("failed: %v\n", err)
			}
		}
		if err := channel.Close(); err != nil {
			log.Printf("failed: %v\n", err)
		}
	}
}

func generateHostKey() (ssh.Signer, error) {
	key, err := generatePrivateKey(privateKeyBitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %s", err.Error())
	}

	private, err := ssh.ParsePrivateKey(encodePrivateKeyToPEM(key))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %s", err.Error())
	}
	return private, nil
}

func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	pk, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	if err = pk.Validate(); err != nil {
		return nil, err
	}

	return pk, nil
}

func encodePrivateKeyToPEM(pk *rsa.PrivateKey) (payload []byte) {
	block := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(pk),
	}
	return pem.EncodeToMemory(&block)
}

func createListener() (net.Listener, error) {
	mrand.Seed(time.Now().Unix())

	var lastErr error

	for i := 0; i < maxAttemptsToCreateListener; i++ {
		var l net.Listener

		l, lastErr = net.Listen("tcp", fmt.Sprintf("localhost:%d", randomPort()))
		if lastErr != nil {
			continue
		}

		return l, nil
	}

	return nil, lastErr
}

func randomPort() int {
	return mrand.Intn(20000-2000) + 2000
}
