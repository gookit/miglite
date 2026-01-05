package command

import (
	"strings"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/x/ccolor"
)

var (
	// Version represents the version of the application
	Version   = "0.1.0"
	BuildTime = "2025-11-05T09:00:00Z"
	GitCommit = "ab3cd4ef"
	GoVersion = "1.21"
)

// SetBuildInfo initializes the version, build time, and git commit
func SetBuildInfo(version, goVer, buildTime, gitCommit string) {
	Version = version
	GoVersion = goVer
	BuildTime = strings.Trim(buildTime, `'"`)
	GitCommit = gitCommit
}

var showVersion bool

// NewApp creates a new CLI application
func NewApp(name, version, description string) *capp.App {
	app := capp.NewWith(name, version, description)

	// Add global flags
	app.BoolVar(&showVersion, "version", false, "Show version and exit;;V")
	app.BoolVar(&ShowVerbose, "verbose", false, "Enable verbose output;;v")
	app.StringVar(&ConfigFile, "config", "./miglite.yaml", "Path to the configuration file;;c")

	// Add commands to the app
	app.Add(
		InitCommand(),
		CreateCommand(),
		NewUpCommand(),
		DownCommand(),
		SkipCommand(),
		StatusCommand(),
		NewExecCommand(),
		NewShowCommand(),
	)

	app.OnAppFlagParsed = beforeRun

	return app
}

func beforeRun(app *capp.App) bool {
	if showVersion {
		ccolor.Printf(`<green>Version</> : %s
<green>Author</>  : https://github.com/inhere
<green>Homepage</>: https://github.com/gookit/miglite

<green>Build Date</>: %s
<green>Git Commit</>: %s
<green>Go Version</>: %s
`, app.Version, BuildTime, strutil.Substr(GitCommit, 0, 10), GoVersion)
		return false
	}

	return true
}
