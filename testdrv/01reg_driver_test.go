package testdrv

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/sysutil"
	"github.com/gookit/miglite/pkg/command"

	// Register database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var app = command.NewApp("miglite", "0.0.1", "Go minimal database migration tool")

func TestMain(m *testing.M) {
	fmt.Println("[TestMain] Workdir:", sysutil.Workdir())
	m.Run()
}
