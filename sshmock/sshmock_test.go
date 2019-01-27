package sshmock

import (
	"testing"

	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

func TestServer(t *testing.T) {
	expectedAnswer := `world`
	command := "hello"
	answers := map[string][]byte{command: []byte(expectedAnswer)}

	s, err := New(answers, true)
	if err != nil {
		t.Fatalf("found errors during setup Server %s", err.Error())
	}
	address := s.Address()

	sshClient, err := sshclient.New(address, "super", "test")
	if err != nil {
		t.Fatalf("unable to connect to ssh server %s", err.Error())
	}

	answer, err := sshClient.Run(command)
	if err != nil {
		t.Fatalf("unable to run command %s: %s", command, err.Error())
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
