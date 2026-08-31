package command

import (
	"fmt"
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
	"strings"
)

type StatusOption struct{}

func StatusCommand() *capp.Cmd {
	c := capp.NewCmd("status", "Show the status of migrations", func(*capp.Cmd) error { return HandleStatus(StatusOption{}) })
	c.Aliases = []string{"st"}
	bindCommonFlags(c)
	return c
}

func HandleStatus(_ StatusOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()
	statuses, err := r.Status(runtime.StatusOption{})
	if err != nil {
		return err
	}
	ccolor.Cyanf("\n📊  Migrations Status:(total=%d)\n", len(statuses))
	fmt.Println(strings.Repeat("==", 44))
	ccolor.Printf("  <b>Status</>  | %13s<b>Version(migration file)</>%13s    |   <b>Operate Time</> \n", "", "")
	fmt.Println(strings.Repeat("--", 44))
	for _, st := range statuses {
		statusIcon := "<mga>pending</>"
		switch st.Status {
		case "up":
			statusIcon = "<green>applied</>"
		case "down":
			statusIcon = "<ylw>rolled</> "
		case "skip":
			statusIcon = "<gray>skipped</>"
		}
		ccolor.Printf("  %s | %-52s | %s\n", statusIcon, st.Version, formatTime(st.AppliedAt))
	}
	return nil
}
