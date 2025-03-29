package charm

import (
	"fmt"
	"os"

	"github.com/gruyaume/certificates-operator/internal/integrations/tls_certificates"
	"github.com/gruyaume/goops/commands"
)

const (
	CaCertificateSecretLabel   = "active-ca-certificates" // #nosec G101
	TLSCertificatesIntegration = "certificates"
)

func isConfigValid(hookCommand *commands.HookCommand) (bool, error) {
	caCommonNameConfig, err := commands.ConfigGet(hookCommand, "ca-common-name")
	if err != nil {
		return false, fmt.Errorf("could not get config: %w", err)
	}
	if caCommonNameConfig == "" {
		return false, fmt.Errorf("ca-common-name config is empty")
	}
	return true, nil
}

func generateAndStoreRootCertificate(hookCommand *commands.HookCommand, logger *commands.Logger) error {
	caCommonName, err := commands.ConfigGet(hookCommand, "ca-common-name")
	if err != nil {
		return fmt.Errorf("could not get config: %w", err)
	}

	_, err = commands.SecretGet(hookCommand, "", CaCertificateSecretLabel, false, true)
	if err != nil {
		logger.Info("could not get secret:", err.Error())
		caCertPEM, caKeyPEM, err := GenerateRootCertificate(caCommonName)
		if err != nil {
			return fmt.Errorf("could not generate root certificate: %w", err)
		}
		logger.Info("Generated new root certificate")
		secretContent := map[string]string{
			"private-key":    caKeyPEM,
			"ca-certificate": caCertPEM,
		}
		_, err = commands.SecretAdd(hookCommand, secretContent, "", CaCertificateSecretLabel)
		if err != nil {
			return fmt.Errorf("could not add secret: %w", err)
		}
		logger.Info("Created new secret")
		return nil
	}
	logger.Info("Secret found")
	return nil
}

func processOutstandingCertificateRequests(hookCommand *commands.HookCommand, logger *commands.Logger) error {
	outstandingCertificateRequests, err := tls_certificates.GetOutstandingCertificateRequests(hookCommand, TLSCertificatesIntegration)
	if err != nil {
		return fmt.Errorf("could not get outstanding certificate requests: %w", err)
	}
	for _, request := range outstandingCertificateRequests {
		logger.Info("Received a certificate signing request from:", request.RelationID, "with common name:", request.CertificateSigningRequest.CommonName)
		caCertificateSecret, err := commands.SecretGet(hookCommand, "", CaCertificateSecretLabel, false, true)
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
		providerCertificatte := tls_certificates.ProviderCertificate{
			RelationID:                request.RelationID,
			Certificate:               tls_certificates.Certificate{Raw: certPEM},
			CertificateSigningRequest: request.CertificateSigningRequest,
			CA:                        tls_certificates.Certificate{Raw: caCertPEM},
			Chain: []tls_certificates.Certificate{
				{Raw: caCertPEM},
			},
			Revoked: false,
		}
		err = tls_certificates.SetRelationCertificate(hookCommand, request.RelationID, providerCertificatte)
		if err != nil {
			logger.Warning("Could not set relation certificate:", err.Error())
			continue
		}
		logger.Info("Provided certificate to:", request.RelationID)
	}
	return nil
}

func HandleDefaultHook(hookCommand *commands.HookCommand, logger *commands.Logger) error {
	isLeader, err := commands.IsLeader(hookCommand)
	if err != nil {
		return fmt.Errorf("could not check if unit is leader: %w", err)
	}
	if !isLeader {
		return fmt.Errorf("unit is not leader")
	}
	valid, err := isConfigValid(hookCommand)
	if err != nil {
		return fmt.Errorf("could not check config: %w", err)
	}
	if !valid {
		return fmt.Errorf("config is not valid")
	}

	err = commands.StatusSet(hookCommand, commands.StatusActive)
	if err != nil {
		logger.Error("Could not set status:", err.Error())
		os.Exit(0)
	}
	logger.Info("Status set to active")

	err = generateAndStoreRootCertificate(hookCommand, logger)
	if err != nil {
		return fmt.Errorf("could not generate CA certificate: %w", err)
	}
	err = processOutstandingCertificateRequests(hookCommand, logger)
	if err != nil {
		return fmt.Errorf("could not process outstanding certificate requests: %w", err)
	}
	return nil
}
