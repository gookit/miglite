package cli

import (
	"fmt"
	"os"

	"github.com/gookit/goutil/cflag"
)

// App represents the CLI application
type App struct {
	Name        string
	Version     string
	Description string
	Commands    map[string]Command
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	ConfigFn    func(*cflag.Command)
	RunFn       func(*cflag.Command) error
}

// NewApp creates a new CLI application
func NewApp(name, version, description string) *App {
	return &App{
		Name:        name,
		Version:     version,
		Description: description,
		Commands:    make(map[string]Command),
	}
}

// AddCommand adds a command to the application
func (app *App) AddCommand(cmd Command) {
	app.Commands[cmd.Name] = cmd
}

// Run executes the CLI application
func (app *App) Run(args []string) error {
	// Create the root command
	rootCmd := cflag.NewCommand("", app.Description)
	rootCmd.Version = app.Version

	// Add global flags
	rootCmd.BoolOpt(&showHelp, "help", "h", false, "Show help message and exit")
	rootCmd.BoolOpt(&showVersion, "version", "v", false, "Show version and exit")

	// Add subcommands
	for _, cmd := range app.Commands {
		subCmd := cflag.NewCommand(cmd.Name, cmd.Description)
		if cmd.ConfigFn != nil {
			cmd.ConfigFn(subCmd)
		}
		subCmd.Run = func(c *cflag.Command) error {
			return cmd.RunFn(c)
		}
		rootCmd.AddCommand(subCmd)
	}

	// Parse and run
	if len(args) == 0 {
		args = os.Args[1:]
	}

	// Check for help or version
	if len(args) > 0 {
		if args[0] == "-h" || args[0] == "--help" {
			showHelp = true
		} else if args[0] == "-v" || args[0] == "--version" {
			showVersion = true
		}
	}

	if showHelp {
		rootCmd.ShowHelp()
		return nil
	}

	if showVersion {
		fmt.Printf("%s version %s\n", app.Name, app.Version)
		return nil
	}

	// Set up the main command's run function
	rootCmd.Run = func(c *cflag.Command) error {
		if len(c.Args()) == 0 {
			c.ShowHelp()
			return nil
		}
		// If we get here, it's an unknown command
		return fmt.Errorf("unknown command: %s", c.Args()[0])
	}

	return rootCmd.ParseAndRun(args)
}

var showHelp bool
var showVersion bool