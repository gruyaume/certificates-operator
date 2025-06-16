package main

import (
	"os"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/goops"
)

func main() {
	env := goops.ReadEnv()

	switch env.HookName {
	case "":
		goops.LogInfof("No hook specified")
	default:
		err := charm.Configure()
		if err != nil {
			goops.LogErrorf("Failed to configure charm: %s", err.Error())
			os.Exit(1)
		}

		goops.LogInfof("Charm configured successfully")
	}

	switch env.ActionName {
	case "get-ca-certificate":
		err := charm.HandleGetCACertificateAction()
		if err != nil {
			goops.LogErrorf("Failed to handle get-ca-certificate action: %s", err.Error())
			os.Exit(1)
		}

	case "":
		goops.LogInfof("No action specified")
	default:
		goops.LogInfof("No action handler for action: %s", env.ActionName)
	}
}
