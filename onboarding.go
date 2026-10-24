package runtimekit

import "fmt"

// OnboardingMode is the first homeowner-facing action after package installation.
type OnboardingMode string

const (
	OnboardingModeManual    OnboardingMode = "manual"
	OnboardingModeDiscovery OnboardingMode = "discovery"
	OnboardingModeAccount   OnboardingMode = "account"
	OnboardingModeService   OnboardingMode = "service"
	OnboardingModeNone      OnboardingMode = "none"
)

// OnboardingDeclaration is the manifest declaration consumed by Core's setup planner.
type OnboardingDeclaration struct {
	Mode OnboardingMode `json:"mode"`
}

// Validate rejects values outside the homeowner-facing onboarding contract.
func (declaration OnboardingDeclaration) Validate() error {
	switch declaration.Mode {
	case OnboardingModeManual,
		OnboardingModeDiscovery,
		OnboardingModeAccount,
		OnboardingModeService,
		OnboardingModeNone:
		return nil
	default:
		return fmt.Errorf(
			"onboarding.mode must be one of: manual, discovery, account, service, none",
		)
	}
}
