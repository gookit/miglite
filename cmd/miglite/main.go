package main

import (
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/pkg/command"

	// Register database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

//
// 需要通过:
// go build -ldflags \
//  "-X 'main.Version=0.0.1' -X 'main.BuildTime=2025-11-05T09:00:00Z' -X 'main.GitCommit=abc3d4efg'" \
//  -o miglite ./cmd/miglite
// 设置下面的信息
//
var (
	// Version represents the version of the application
	Version   = "0.1.0"
	BuildTime = "2025-11-05 09:00:00"
	GitCommit = "ab3cd4ef"
	GoVersion = "go1.20.1" // go version on build
)

// DEV
//  run: go run .\cmd\miglite
//  install: go install .\cmd\miglite
func main() {
	miglite.InitInfo(Version, GoVersion, BuildTime, GitCommit)

	// Create the CLI application
	app := command.NewApp("miglite", Version, "Go minimal database migration tool")

	// Run the application
	app.Run()
	// For testing
	// err := app.RunWithArgs([]string{"init", "-h"})
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
