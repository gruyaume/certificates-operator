package charm

import (
	"context"
	"fmt"

	"github.com/gruyaume/charm-libraries/certificates"
	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"go.opentelemetry.io/otel"
)

const (
	CaCertificateSecretLabel   = "active-ca-certificates" // #nosec G101
	TLSCertificatesIntegration = "certificates"
)

func generateAndStoreRootCertificate(hookContext *goops.HookContext) error {
	config := &ConfigOptions{}

	err := config.LoadFromJuju(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	secretGetOpts := &commands.SecretGetOptions{
		Label:   CaCertificateSecretLabel,
		Refresh: true,
	}

	_, err = hookContext.Commands.SecretGet(secretGetOpts)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Info, "could not get secret:", err.Error())

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

		hookContext.Commands.JujuLog(commands.Info, "Generated new root certificate")

		secretContent := map[string]string{
			"private-key":    caKeyPEM,
			"ca-certificate": caCertPEM,
		}
		secretAddOpts := &commands.SecretAddOptions{
			Label:   CaCertificateSecretLabel,
			Content: secretContent,
		}

		_, err = hookContext.Commands.SecretAdd(secretAddOpts)
		if err != nil {
			return fmt.Errorf("could not add secret: %w", err)
		}

		hookContext.Commands.JujuLog(commands.Info, "Created new secret")

		return nil
	}

	hookContext.Commands.JujuLog(commands.Info, "Secret found")

	return nil
}

func processOutstandingCertificateRequests(hookContext *goops.HookContext) error {
	tlsProvider := &certificates.IntegrationProvider{
		RelationName: TLSCertificatesIntegration,
		HookContext:  hookContext,
	}

	outstandingCertificateRequests, err := tlsProvider.GetOutstandingCertificateRequests()
	if err != nil {
		return fmt.Errorf("could not get outstanding certificate requests: %w", err)
	}

	for _, request := range outstandingCertificateRequests {
		hookContext.Commands.JujuLog(commands.Info, "Received a certificate signing request from:", request.RelationID, "with common name:", request.CertificateSigningRequest.CommonName)

		secretGetOpts := &commands.SecretGetOptions{
			Label:   CaCertificateSecretLabel,
			Refresh: true,
		}

		caCertificateSecret, err := hookContext.Commands.SecretGet(secretGetOpts)
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
			hookContext.Commands.JujuLog(commands.Warning, "Could not set relation certificate:", err.Error())
			continue
		}

		hookContext.Commands.JujuLog(commands.Info, "Provided certificate to:", request.RelationID)
	}

	return nil
}

func HandleDefaultHook(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("certificates").Start(ctx, "Handle Hook")
	defer span.End()

	isLeader, err := hookContext.Commands.IsLeader()
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not check if unit is leader", err.Error())
		return
	}

	if !isLeader {
		hookContext.Commands.JujuLog(commands.Info, "Unit is not leader")
		return
	}

	err = validateConfigOptions(ctx, hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Config validation failed:", err.Error())
		return
	}

	err = generateAndStoreRootCertificate(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not generate and store root certificate:", err.Error())
		return
	}

	err = processOutstandingCertificateRequests(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not process outstanding certificate requests:", err.Error())
		return
	}
}

func SetStatus(ctx context.Context, hookContext *goops.HookContext) {
	_, span := otel.Tracer("certificates").Start(ctx, "Set Status")
	defer span.End()

	status := commands.StatusActive

	message := ""

	err := validateConfigOptions(ctx, hookContext)
	if err != nil {
		status = commands.StatusBlocked
		message = fmt.Sprintf("Invalid config: %s", err.Error())
	}

	statusSetOpts := &commands.StatusSetOptions{
		Name:    status,
		Message: message,
	}

	err = hookContext.Commands.StatusSet(statusSetOpts)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Error, "Could not set status:", err.Error())
		return
	}
}

func validateConfigOptions(ctx context.Context, hookContext *goops.HookContext) error {
	_, span := otel.Tracer("certificates").Start(ctx, "validate config")
	defer span.End()

	config := &ConfigOptions{}

	err := config.LoadFromJuju(hookContext)
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Couldn't load config options: %s", err.Error())
		return fmt.Errorf("couldn't load config options: %w", err)
	}

	err = config.Validate()
	if err != nil {
		hookContext.Commands.JujuLog(commands.Warning, "Config is not valid: %s", err.Error())
		return fmt.Errorf("config is not valid: %w", err)
	}

	return nil
}
