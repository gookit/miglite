package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
)

var showVersion bool

// NewApp creates a new CLI application
func NewApp(name, version, description string) *capp.App {
	app := capp.NewWith(name, version, description)

	// Add global flags
	// app.BoolVar(&showHelp, "help", false, "Show help message and exit;;h")
	app.BoolVar(&showVersion, "version", false, "Show version and exit;;V")
	app.BoolVar(&showVerbose, "verbose", false, "Enable verbose output;;v")

	// Add commands to the app
	app.Add(
		NewUpCommand(),
		DownCommand(),
		StatusCommand(),
		CreateCommand(),
		InitCommand(),
	)

	app.OnAppFlagParsed = beforeRun

	return app
}

func beforeRun(app *capp.App) bool {
	if showVersion {
		ccolor.Printf(`<green>Version</> : v%s
<green>Author</>  : https://github.com/inhere
<green>Homepage</>: https://github.com/gookit/miglite
`, app.Version)
		return false
	}

	return true
}
