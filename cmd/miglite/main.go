package main

import (
	"os"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/miglite/cmd/miglite/commands"
	"github.com/gookit/miglite/pkg/cli"
)

func main() {
	// Create the CLI application
	app := cli.NewApp("miglite", "1.0.0", "Go minimal database migration tool")
	
	// Add commands to the app
	app.AddCommand(cli.Command{
		Name:        "up",
		Description: "Run migrations",
		ConfigFn: func(cmd *cflag.Command) {
			cmd.BoolOpt(nil, "verbose", "v", false, "Enable verbose output")
			cmd.StringOpt(nil, "config", "c", "./miglite.yaml", "Configuration file path")
		},
		RunFn: commands.UpCommand(),
	})
	
	app.AddCommand(cli.Command{
		Name:        "status",
		Description: "Show migrations status",
		ConfigFn: func(cmd *cflag.Command) {
			cmd.StringOpt(nil, "config", "c", "./miglite.yaml", "Configuration file path")
		},
		RunFn: commands.StatusCommand(),
	})
	
	app.AddCommand(cli.Command{
		Name:        "down",
		Description: "Rollback migrations",
		ConfigFn: func(cmd *cflag.Command) {
			cmd.IntOpt(nil, "count", "", 1, "Number of migrations to rollback")
			cmd.StringOpt(nil, "config", "c", "./miglite.yaml", "Configuration file path")
		},
		RunFn: commands.DownCommand(),
	})
	
	app.AddCommand(cli.Command{
		Name:        "create",
		Description: "Create a new migration",
		ConfigFn: func(cmd *cflag.Command) {
			cmd.StringOpt(nil, "config", "c", "./miglite.yaml", "Configuration file path")
		},
		RunFn: commands.CreateCommand(),
	})

	// Run the application
	if err := app.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}