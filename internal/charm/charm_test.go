package charm_test

import (
	"testing"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/goops/goopstest"
)

func TestGivenNotLeaderWhenConfigureThenBlockedStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := goopstest.State{
		Leader: false,
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusBlocked,
		Message: "Unit is not leader",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected unit status %s, got %s", expectedStatus.Name, stateOut.UnitStatus)
	}
}

func TestGivenInvalidConfigWhenConfigureThenBlockedStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"ca-common-name": "",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusBlocked,
		Message: "Invalid config: config is not valid: ca-common-name config is empty",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected unit status %s, got %s", expectedStatus.Name, stateOut.UnitStatus)
	}
}

func TestGivenGoodConfigWhenConfigureThenActiveStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"ca-common-name": "pizza.com",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusActive,
		Message: "Certificates operator is running",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected unit status %s, got %s", expectedStatus.Name, stateOut.UnitStatus)
	}
}

func TestGivenGoodConfigWhenConfigureThenPrivateKeySecretCreated(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := goopstest.State{
		Leader: true,
		Config: map[string]any{
			"ca-common-name": "pizza.com",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stateOut.Secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(stateOut.Secrets))
	}

	if stateOut.Secrets[0].Label != "active-ca-certificates" {
		t.Errorf("expected secret label 'active-ca-certificates', got '%s'", stateOut.Secrets[0].Label)
	}

	if _, ok := stateOut.Secrets[0].Content["private-key"]; !ok {
		t.Errorf("expected secret to contain 'private-key', but it was not found")
	}

	if _, ok := stateOut.Secrets[0].Content["ca-certificate"]; !ok {
		t.Errorf("expected secret to contain 'ca-certificate', but it was not found")
	}
}
