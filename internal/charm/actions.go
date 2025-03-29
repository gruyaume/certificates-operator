package charm

import (
	"fmt"

	"github.com/gruyaume/go-operator/commands"
)

func HandleGetCACertificateAction(commandRunner *commands.DefaultRunner) error {
	caCertificateSecret, err := commands.SecretGet(commandRunner, "", CaCertificateSecretLabel, false, true)
	if err != nil {
		err := commands.ActionFail(commandRunner, "could not get CA certificate secret")
		if err != nil {
			return fmt.Errorf("could not fail action: %w and could not get CA certificate secret: %w", err, err)
		}
		return fmt.Errorf("could not get CA certificate secret: %w", err)
	}
	caCertPEM, ok := caCertificateSecret["ca-certificate"]
	if !ok {
		err := commands.ActionFail(commandRunner, "could not find CA certificate in secret")
		if err != nil {
			return fmt.Errorf("could not fail action: %w and could not find CA certificate in secret: %w", err, err)
		}
		return fmt.Errorf("could not find CA certificate in secret")
	}
	err = commands.ActionSet(commandRunner, map[string]string{"ca-certificate": caCertPEM})
	if err != nil {
		err := commands.ActionFail(commandRunner, "could not set action result")
		if err != nil {
			return fmt.Errorf("could not fail action: %w and could not set action result: %w", err, err)
		}
		return fmt.Errorf("could not set action result: %w", err)
	}
	return nil
}
