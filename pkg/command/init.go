package command

import (
	"fmt"

	"github.com/gookit/goutil/cflag"
	"github.com/gookit/goutil/x/ccolor"
)

// InitCommand initializes the migration schema on db
func InitCommand() *cflag.Cmd {
	c := cflag.NewCmd("init", "Initialize the migration schema on db")

	c.Func = func(c *cflag.Cmd) error {
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
