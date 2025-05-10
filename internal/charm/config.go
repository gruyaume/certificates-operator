package charm

import (
	"fmt"

	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
)

type ConfigOptions struct {
	caCommonName          string
	caOrganization        string
	caOrganizationalUnit  string
	caEmailAddress        string
	caCountryName         string
	caStateOrProvinceName string
	caLocalityName        string
}

func (c *ConfigOptions) LoadFromJuju(hookContext *goops.HookContext) error {
	caCommonName, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-common-name"})
	if err != nil {
		return fmt.Errorf("failed to get ca-common-name config: %w", err)
	}

	caOrganization, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-organization"})
	if err != nil {
		return fmt.Errorf("failed to get ca-organization config: %w", err)
	}

	caOrganizationalUnit, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-organizational-unit"})
	if err != nil {
		return fmt.Errorf("failed to get ca-organizational-unit config: %w", err)
	}

	caEmailAddress, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-email-address"})
	if err != nil {
		return fmt.Errorf("failed to get ca-email-address config: %w", err)
	}

	caCountryName, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-country-name"})
	if err != nil {
		return fmt.Errorf("failed to get ca-country-name config: %w", err)
	}

	caStateOrProvinceName, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-state-or-province-name"})
	if err != nil {
		return fmt.Errorf("failed to get ca-state-or-province-name config: %w", err)
	}

	caLocalityName, err := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-locality-name"})
	if err != nil {
		return fmt.Errorf("failed to get ca-locality-name config: %w", err)
	}

	c.caCommonName = caCommonName
	c.caOrganization = caOrganization
	c.caOrganizationalUnit = caOrganizationalUnit
	c.caEmailAddress = caEmailAddress
	c.caCountryName = caCountryName
	c.caStateOrProvinceName = caStateOrProvinceName
	c.caLocalityName = caLocalityName

	return nil
}

func (c *ConfigOptions) Validate() error {
	if c.caCommonName == "" {
		return fmt.Errorf("ca-common-name config is empty")
	}

	return nil
}
