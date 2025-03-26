package sshclient

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	clientTimeout = 15 * time.Second
	sshPort       = "22"
)

// SSHClient implements out common abstraction for SSH
type SSHClient struct {
	addr   string
	config *ssh.ClientConfig
	client *ssh.Client
	lock   *sync.Mutex
}

// New creates a new SSH client
func New(addr, username, password string) (*SSHClient, error) {
	cfg := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				if len(questions) == 0 {
					return []string{}, nil
				}
				if len(questions) == 1 {
					return []string{password}, nil
				}
				return []string{}, fmt.Errorf("unsupported keyboard-interactive auth")
			}),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: clientTimeout,
	}

	addr, err := checkAndBuildAddr(addr)
	if err != nil {
		return nil, err
	}

	return &SSHClient{addr: addr, config: cfg, lock: new(sync.Mutex)}, nil
}

// Run executes the given command and returns the output as a string
func (s *SSHClient) Run(command string) (string, error) {
	if err := s.createClient(); err != nil {
		return "", err
	}

	return s.run(command)
}

func (s *SSHClient) run(command string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// Close sends "exit" command and closes the SSH connection
func (s *SSHClient) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.client == nil {
		return nil
	}
	defer func() {
		s.client.Close()
		s.client = nil
	}()

	// some vendors have issues with the bmc if you don't do it
	if _, err := s.run("exit"); err != nil {
		return err
	}

	return nil
}

func (s *SSHClient) createClient() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.client != nil {
		return nil // TODO: check client is alive
	}

	c, err := ssh.Dial("tcp", s.addr, s.config)
	if err != nil {
		return fmt.Errorf("unable to connect to bmc: %w", err)
	}
	s.client = c

	return nil
}

func checkAndBuildAddr(addr string) (string, error) {
	if addr == "" {
		return "", fmt.Errorf("address is empty")
	}

	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr, nil
	}

	addrWithPort := net.JoinHostPort(addr, sshPort)
	if _, _, err := net.SplitHostPort(addrWithPort); err == nil {
		return addrWithPort, nil
	}

	return "", fmt.Errorf("failed to parse address %q", addr)
}
