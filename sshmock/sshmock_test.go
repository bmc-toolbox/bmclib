package sshmock

import (
	"testing"

	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

func Test_Server(t *testing.T) {
	expectedAnswer := "world"
	command := "hello"
	answers := map[string][]byte{
		command: []byte(expectedAnswer),
		"exit":  []byte("see you"),
	}

	s, err := New(answers, nil)
	if err != nil {
		t.Fatalf(err.Error())
	}

	shutdown, address, err := s.ListenAndServe()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer shutdown()

	sshClient, err := sshclient.New(address, "super", "test")
	if err != nil {
		t.Fatalf("unable to connect to ssh server: %v", err)
	}

	answer, err := sshClient.Run(command)
	if err != nil {
		t.Fatalf("unable to run command %s: %v", command, err)
	}

	if answer != expectedAnswer {
		t.Errorf("expected answer %v: found %v", expectedAnswer, answer)
	}

	if err := sshClient.Close(); err != nil {
		t.Errorf("Close() returns an error:%v", err)
	}
}
