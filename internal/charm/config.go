package charm

import (
	"fmt"

	"github.com/gruyaume/goops"
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

func (c *ConfigOptions) LoadFromJuju() error {
	caCommonName, _ := goops.GetConfigString("ca-common-name")

	caOrganization, _ := goops.GetConfigString("ca-organization")

	caOrganizationalUnit, _ := goops.GetConfigString("ca-organizational-unit")

	caEmailAddress, _ := goops.GetConfigString("ca-email-address")

	caCountryName, _ := goops.GetConfigString("ca-country-name")

	caStateOrProvinceName, _ := goops.GetConfigString("ca-state-or-province-name")

	caLocalityName, _ := goops.GetConfigString("ca-locality-name")

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
