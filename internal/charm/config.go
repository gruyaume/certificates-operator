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
	caCommonName, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-common-name"})

	caOrganization, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-organization"})

	caOrganizationalUnit, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-organizational-unit"})

	caEmailAddress, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-email-address"})

	caCountryName, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-country-name"})

	caStateOrProvinceName, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-state-or-province-name"})

	caLocalityName, _ := hookContext.Commands.ConfigGetString(&commands.ConfigGetOptions{Key: "ca-locality-name"})

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
