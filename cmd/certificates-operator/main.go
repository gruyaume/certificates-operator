package main

import (
	"os"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/go-operator/commands"
	"github.com/gruyaume/go-operator/environment"
)

func main() {
	commandRunner := &commands.DefaultRunner{}
	environmentGetter := &environment.DefaultEnvironment{}
	logger := commands.NewLogger(commandRunner)
	actionName := environment.JujuActionName(environmentGetter)
	if actionName != "" {
		logger.Info("Action name:", actionName)
		switch actionName {
		case "get-ca-certificate":
			err := charm.HandleGetCACertificateAction(commandRunner)
			if err != nil {
				logger.Error("Error handling get-ca-certificate action:", err.Error())
				os.Exit(0)
			}
			logger.Info("Handled get-ca-certificate action successfully")
			os.Exit(0)
		default:
			logger.Info("Action not recognized, exiting")
			os.Exit(0)
		}
	}

	hookName := environment.JujuHookName(environmentGetter)
	if hookName != "" {
		logger.Info("Hook name:", hookName)
		err := charm.HandleDefaultHook(commandRunner, logger)
		if err != nil {
			logger.Error("Error handling default hook:", err.Error())
			os.Exit(0)
		}
	}
}
