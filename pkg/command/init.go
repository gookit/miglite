package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
)

// InitCommand initializes the migration schema on db
func InitCommand() *capp.Cmd {
	c := capp.NewCmd("init", "Initialize the migration schema on db")

	c.Func = func(c *capp.Cmd) error {
		_, db, err := initConfigAndDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Initialize schema if needed
		if err := db.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %v", err)
		}

		ccolor.Infoln("Migration schema initialized successfully.")
		return nil
	}
	return c
}
