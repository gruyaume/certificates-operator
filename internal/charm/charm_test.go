package charm_test

import (
	"testing"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/goopstest"
)

func TestGivenNotLeaderWhenConfigureThenBlockedStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: false,
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateOut.UnitStatus != string(goops.StatusBlocked) {
		t.Errorf("expected unit status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenInvalidConfigWhenConfigureThenBlockedStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
			"ca-common-name": "",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateOut.UnitStatus != string(goops.StatusBlocked) {
		t.Errorf("expected unit status %s, got %s", goops.StatusBlocked, stateOut.UnitStatus)
	}
}

func TestGivenGoodConfigWhenConfigureThenActiveStatus(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
			"ca-common-name": "pizza.com",
		},
	}

	stateOut, err := ctx.Run("start", stateIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stateOut.UnitStatus != string(goops.StatusActive) {
		t.Errorf("expected unit status %s, got %s", goops.StatusActive, stateOut.UnitStatus)
	}
}

func TestGivenGoodConfigWhenConfigureThenPrivateKeySecretCreated(t *testing.T) {
	ctx := goopstest.Context{
		Charm: charm.Configure,
	}

	stateIn := &goopstest.State{
		Leader: true,
		Config: map[string]string{
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
