package runtimekit

import (
	"fmt"
	"os"
	"strconv"
)

const RuntimePortEnvName = "PORT"

// RuntimePort resolves the listener port assigned by Core. The default is used
// only when the process is running outside the managed Core environment.
func RuntimePort(defaultPort int) (int, error) {
	if err := validateRuntimePort(defaultPort, "default port"); err != nil {
		return 0, err
	}
	raw, present := os.LookupEnv(RuntimePortEnvName)
	if !present {
		return defaultPort, nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("PORT must be an integer between 1 and 65535: %w", err)
	}
	if err := validateRuntimePort(port, RuntimePortEnvName); err != nil {
		return 0, err
	}
	return port, nil
}

func validateRuntimePort(port int, source string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be an integer between 1 and 65535", source)
	}
	return nil
}
