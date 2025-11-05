package main

import (
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/pkg/command"

	// Register database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func main() {
	// Create the CLI application
	app := command.NewApp("miglite", miglite.Version, "Go minimal database migration tool")

	// Run the application
	app.Run()
}
