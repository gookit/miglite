package main

import (
	"github.com/gookit/miglite/pkg/command"

	// Register database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func main() {
	// Create the CLI application
	app := command.NewApp("miglite", "1.0.0", "Go minimal database migration tool")

	// Add commands to the app
	app.Add(
		command.NewUpCommand(),
		command.StatusCommand(),
		command.CreateCommand(),
		command.DownCommand(),
	)

	// Run the application
	app.Run()
}
