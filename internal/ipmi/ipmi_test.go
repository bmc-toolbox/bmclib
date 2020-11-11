package ipmi

import (
	"context"
	"testing"
	"time"
)

func TestInfo(t *testing.T) {
	t.Skip("needs ipmitool")
	i, err := New("admin", "admin", "128.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	expected := "error getting bmc info: context deadline exceeded"
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err = i.Info(ctx)
	if err == nil {
		t.Fatalf("expected: %v, got: %v", expected, nil)
	}
	if err.Error() != expected {
		t.Fatalf("expected: %v, got: %v", expected, err.Error())
	}
}
