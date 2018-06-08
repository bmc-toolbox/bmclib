package sshclient

import (
	"fmt"
	"net"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh"
)

const (
	// PowerOff defines the action of powering off a device
	PowerOff = "poweroff"
	// PowerOn defines the action of powering on a device
	PowerOn = "poweron"
	// PowerCycle defines the action of power cycle a device
	PowerCycle = "powercycle"
	// HardReset defines the action of hard reset a device
	HardReset = "hardreset"
	// Reseat defines the action of power reseat a device
	Reseat = "reseat"
	// IsOn defines the current power status of a device
	IsOn = "ison"
	// PowerCycleBmc  the action of power cycle the bmc of a device
	PowerCycleBmc = "powercyclebmc"
	// PxeOnce  the action of pxe once a device
	PxeOnce = "pxeonce"
)

// SSHClient implements out commom abstraction for ssh
type SSHClient struct {
	client *ssh.Client
}

// Sleep transforms a sleep statement in a sleep-able time
func Sleep(sleep string) (err error) {
	sleep = strings.Replace(sleep, "sleep ", "", 1)
	s, err := time.ParseDuration(sleep)
	if err != nil {
		return fmt.Errorf("error sleeping: %v", err)
	}
	time.Sleep(s)

	return err
}

// Run execute the given command and returns a string with the output
func (s *SSHClient) Run(command string) (result string, err error) {
	session, err := s.client.NewSession()
	if err != nil {
		return result, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), err
	}

	return string(output), err
}

// IsntLetterOrNumber check if the give rune is not a letter nor a number
func IsntLetterOrNumber(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c)
}

// New returns a new configured ssh client
func New(host string, username string, password string) (connection *SSHClient, err error) {
	c, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:22", host),
		&ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: 15 * time.Second,
		},
	)
	if err != nil {
		return connection, fmt.Errorf("unable to connect to bmc: %v", err)
	}
	return &SSHClient{c}, err
}

// Close closed the ssh connection and ensure to always exit, some vendors will have issues with the bmc if you dont do it
func (s *SSHClient) Close() (err error) {
	defer s.client.Close()
	_, err = s.Run("exit")
	return err
}
