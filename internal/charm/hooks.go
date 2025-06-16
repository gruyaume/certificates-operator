package charm

import (
	"fmt"

	"github.com/gruyaume/charm-libraries/certificates"
	"github.com/gruyaume/goops"
)

const (
	CaCertificateSecretLabel   = "active-ca-certificates" // #nosec G101
	TLSCertificatesIntegration = "certificates"
)

func generateAndStoreRootCertificate() error {
	config := ConfigOptions{}

	err := goops.GetConfig(&config)
	if err != nil {
		return fmt.Errorf("could not get config options: %w", err)
	}

	_, err = goops.GetSecretByLabel(CaCertificateSecretLabel, false, true)
	if err != nil {
		goops.LogInfof("could not get secret: %s", err.Error())

		caCertPEM, caKeyPEM, err := GenerateRootCertificate(&CaCertificateOpts{
			CommonName:          config.CACommonName,
			Organization:        config.CAOrganization,
			OrganizationalUnit:  config.CAOrganizationalUnit,
			EmailAddress:        config.CAEmailAddress,
			CountryName:         config.CACountryName,
			StateOrProvinceName: config.CAStateOrProvinceName,
			LocalityName:        config.CALocalityName,
		})
		if err != nil {
			return fmt.Errorf("could not generate root certificate: %w", err)
		}

		goops.LogInfof("Generated new root certificate with common name: %s", config.CACommonName)

		secretContent := map[string]string{
			"private-key":    caKeyPEM,
			"ca-certificate": caCertPEM,
		}

		_, err = goops.AddSecret(&goops.AddSecretOptions{
			Label:   CaCertificateSecretLabel,
			Content: secretContent,
		})
		if err != nil {
			return fmt.Errorf("could not add secret: %w", err)
		}

		goops.LogInfof("Created new secret")

		return nil
	}

	goops.LogInfof("Secret found")

	return nil
}

func processOutstandingCertificateRequests() error {
	tlsProvider := &certificates.IntegrationProvider{
		RelationName: TLSCertificatesIntegration,
	}

	outstandingCertificateRequests, err := tlsProvider.GetOutstandingCertificateRequests()
	if err != nil {
		return fmt.Errorf("could not get outstanding certificate requests: %w", err)
	}

	for _, request := range outstandingCertificateRequests {
		goops.LogInfof("Received a certificate signing request from: %s with common name: %s", request.RelationID, request.CertificateSigningRequest.CommonName)

		caCertificateSecret, err := goops.GetSecretByLabel(CaCertificateSecretLabel, false, true)
		if err != nil {
			return fmt.Errorf("could not get CA certificate secret: %w", err)
		}

		caKeyPEM, ok := caCertificateSecret["private-key"]
		if !ok {
			return fmt.Errorf("could not find CA private key in secret")
		}

		caCertPEM, ok := caCertificateSecret["ca-certificate"]
		if !ok {
			return fmt.Errorf("could not find CA certificate in secret")
		}

		certPEM, err := GenerateCertificate(caKeyPEM, caCertPEM, request.CertificateSigningRequest.Raw)
		if err != nil {
			return fmt.Errorf("could not generate certificate: %w", err)
		}

		err = tlsProvider.SetRelationCertificate(&certificates.SetRelationCertificateOptions{
			RelationID:                request.RelationID,
			Certificate:               certPEM,
			CA:                        caCertPEM,
			Chain:                     []string{caCertPEM},
			CertificateSigningRequest: request.CertificateSigningRequest.Raw,
		})
		if err != nil {
			goops.LogWarningf("Could not set relation certificate for %s: %s", request.RelationID, err.Error())
			continue
		}

		goops.LogInfof("Provided certificate to: %s", request.RelationID)
	}

	return nil
}

func Configure() error {
	isLeader, err := goops.IsLeader()
	if err != nil {
		return fmt.Errorf("could not check if unit is leader: %w", err)
	}

	if !isLeader {
		_ = goops.SetUnitStatus(goops.StatusBlocked, "Unit is not leader")
		return nil
	}

	err = validateConfigOptions()
	if err != nil {
		_ = goops.SetUnitStatus(goops.StatusBlocked, fmt.Sprintf("Invalid config: %s", err.Error()))
		return nil
	}

	err = generateAndStoreRootCertificate()
	if err != nil {
		return fmt.Errorf("could not generate and store root certificate: %w", err)
	}

	err = processOutstandingCertificateRequests()
	if err != nil {
		return fmt.Errorf("could not process outstanding certificate requests: %w", err)
	}

	_ = goops.SetUnitStatus(goops.StatusActive, "Certificates operator is running")

	return nil
}

func validateConfigOptions() error {
	config := ConfigOptions{}

	err := goops.GetConfig(&config)
	if err != nil {
		return fmt.Errorf("could not get config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return fmt.Errorf("config is not valid: %w", err)
	}

	return nil
}
