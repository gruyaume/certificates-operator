package charm

import (
	"context"
	"fmt"

	"github.com/gruyaume/charm-libraries/certificates"
	"github.com/gruyaume/goops"
	"go.opentelemetry.io/otel"
)

const (
	CaCertificateSecretLabel   = "active-ca-certificates" // #nosec G101
	TLSCertificatesIntegration = "certificates"
)

func generateAndStoreRootCertificate() error {
	config := &ConfigOptions{}

	err := config.LoadFromJuju()
	if err != nil {
		goops.LogWarningf("Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	_, err = goops.GetSecretByLabel(CaCertificateSecretLabel, false, true)
	if err != nil {
		goops.LogInfof("could not get secret: %s", err.Error())

		caCertPEM, caKeyPEM, err := GenerateRootCertificate(&CaCertificateOpts{
			CommonName:          config.caCommonName,
			Organization:        config.caOrganization,
			OrganizationalUnit:  config.caOrganizationalUnit,
			EmailAddress:        config.caEmailAddress,
			CountryName:         config.caCountryName,
			StateOrProvinceName: config.caStateOrProvinceName,
			LocalityName:        config.caLocalityName,
		})
		if err != nil {
			return fmt.Errorf("could not generate root certificate: %w", err)
		}

		goops.LogInfof("Generated new root certificate")

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

func HandleDefaultHook(ctx context.Context) {
	_, span := otel.Tracer("certificates").Start(ctx, "Handle Hook")
	defer span.End()

	isLeader, err := goops.IsLeader()
	if err != nil {
		goops.LogErrorf("Could not check if unit is leader: %s", err.Error())
		return
	}

	if !isLeader {
		goops.LogInfof("Unit is not leader, skipping hook execution")
		return
	}

	err = validateConfigOptions(ctx)
	if err != nil {
		goops.LogErrorf("Config validation failed: %s", err.Error())
		return
	}

	err = generateAndStoreRootCertificate()
	if err != nil {
		goops.LogErrorf("Could not generate and store root certificate: %s", err.Error())
		return
	}

	err = processOutstandingCertificateRequests()
	if err != nil {
		goops.LogErrorf("Could not process outstanding certificate requests: %s", err.Error())
		return
	}
}

func SetStatus(ctx context.Context) {
	_, span := otel.Tracer("certificates").Start(ctx, "Set Status")
	defer span.End()

	status := goops.StatusActive

	message := ""

	err := validateConfigOptions(ctx)
	if err != nil {
		status = goops.StatusBlocked
		message = fmt.Sprintf("Invalid config: %s", err.Error())
	}

	err = goops.SetUnitStatus(status, message)
	if err != nil {
		goops.LogErrorf("Could not set status: %s", err.Error())
		return
	}
}

func validateConfigOptions(ctx context.Context) error {
	_, span := otel.Tracer("certificates").Start(ctx, "validate config")
	defer span.End()

	config := &ConfigOptions{}

	err := config.LoadFromJuju()
	if err != nil {
		goops.LogWarningf("Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		goops.LogWarningf("Config is not valid: %s", err.Error())
		return fmt.Errorf("config is not valid: %w", err)
	}

	return nil
}
