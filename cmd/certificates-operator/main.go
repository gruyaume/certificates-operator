package main

import (
	"os"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/goops/commands"
	"github.com/gruyaume/goops/environment"
)

func main() {
	hookCommand := &commands.HookCommand{}
	execEnv := &environment.ExecutionEnvironment{}
	logger := commands.NewLogger(hookCommand)
	actionName := environment.JujuActionName(execEnv)
	if actionName != "" {
		logger.Info("Action name:", actionName)
		switch actionName {
		case "get-ca-certificate":
			err := charm.HandleGetCACertificateAction(hookCommand)
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

	hookName := environment.JujuHookName(execEnv)
	if hookName != "" {
		logger.Info("Hook name:", hookName)
		err := charm.HandleDefaultHook(hookCommand, logger)
		if err != nil {
			logger.Error("Error handling default hook:", err.Error())
			os.Exit(0)
		}
	}
}
