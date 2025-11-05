package command

import (
	"github.com/gookit/goutil/cflag"
)

var showVersion bool

// NewApp creates a new CLI application
func NewApp(name, version, description string) *cflag.App {
	app := cflag.NewApp()
	app.Name = name
	app.Version = version
	app.Desc = description

	// Add global flags
	// app.BoolVar(&showHelp, "help", false, "Show help message and exit;;h")
	app.BoolVar(&showVersion, "version", false, "Show version and exit;;v")

	// Add commands to the app
	app.Add(
		StatusCommand(),
		CreateCommand(),
		DownCommand(),
		NewUpCommand(),
	)

	return app
}
