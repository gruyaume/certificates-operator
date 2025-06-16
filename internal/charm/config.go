package charm

import (
	"fmt"
)

type ConfigOptions struct {
	CACommonName          string `json:"ca-common-name"`
	CAOrganization        string `json:"ca-organization"`
	CAOrganizationalUnit  string `json:"ca-organizational-unit"`
	CAEmailAddress        string `json:"ca-email-address"`
	CACountryName         string `json:"ca-country-name"`
	CAStateOrProvinceName string `json:"ca-state-or-province-name"`
	CALocalityName        string `json:"ca-locality-name"`
}

func (c *ConfigOptions) Validate() error {
	if c.CACommonName == "" {
		return fmt.Errorf("ca-common-name config is empty")
	}

	return nil
}
