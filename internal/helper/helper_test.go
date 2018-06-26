package helper

import "testing"

func TestWhosCalling(t *testing.T) {
	expectedAnswer := "TestWhosCalling"

	answer := WhosCalling()

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
