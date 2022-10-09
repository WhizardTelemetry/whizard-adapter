package main

import (
	"os"

	"github.com/WhizardTelemetry/whizard-adapter/cmd/app"
)

func main() {
	command := app.NewCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
