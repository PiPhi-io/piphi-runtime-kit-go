package runtimekit

import (
	"os"
	"testing"
)

func TestRuntimePortFallsBackOutsideCore(t *testing.T) {
	t.Setenv(RuntimePortEnvName, "43127")
	port, err := RuntimePort(8090)
	if err != nil || port != 43127 {
		t.Fatalf("RuntimePort() = %d, %v; want 43127, nil", port, err)
	}
}

func TestRuntimePortUsesDefaultWhenUnset(t *testing.T) {
	t.Setenv(RuntimePortEnvName, "43127")
	if err := os.Unsetenv(RuntimePortEnvName); err != nil {
		t.Fatal(err)
	}
	port, err := RuntimePort(8090)
	if err != nil || port != 8090 {
		t.Fatalf("RuntimePort() = %d, %v; want 8090, nil", port, err)
	}
}

func TestRuntimePortRejectsInvalidAssignment(t *testing.T) {
	for _, value := range []string{"", "nope", "0", "65536"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(RuntimePortEnvName, value)
			if _, err := RuntimePort(8090); err == nil {
				t.Fatal("RuntimePort() error = nil; want contract error")
			}
		})
	}
}
