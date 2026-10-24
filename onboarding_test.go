package runtimekit

import "testing"

func TestOnboardingDeclarationAcceptsCanonicalModes(t *testing.T) {
	modes := []OnboardingMode{
		OnboardingModeManual,
		OnboardingModeDiscovery,
		OnboardingModeAccount,
		OnboardingModeService,
		OnboardingModeNone,
	}
	for _, mode := range modes {
		if err := (OnboardingDeclaration{Mode: mode}).Validate(); err != nil {
			t.Fatalf("expected %q to be valid: %v", mode, err)
		}
	}
}

func TestOnboardingDeclarationRejectsRuntimeDiscoveryStrategy(t *testing.T) {
	err := (OnboardingDeclaration{Mode: OnboardingMode("passive")}).Validate()
	if err == nil {
		t.Fatal("expected passive runtime discovery to be rejected as onboarding")
	}
}
