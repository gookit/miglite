package command

import (
	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/x/ccolor"
	"github.com/gookit/miglite/internal/runtime"
)

type InitOption struct {
	Drop bool
}

// InitCommand initializes the migration schema on db
func InitCommand() *capp.Cmd {
	var initOpt = InitOption{}
	c := capp.NewCmd("init", "Initialize the migration schema on database")

	bindCommonFlags(c)
	c.BoolVar(&initOpt.Drop, "drop", false, "Drop existing schema before create")

	c.Func = func(c *capp.Cmd) error {
		return HandleInit(initOpt)
	}
	return c
}

// HandleInit handles the init command logic
func HandleInit(opt InitOption) error {
	r, cleanup, err := legacyRuntime()
	if err != nil {
		return err
	}
	defer cleanup()

	err = r.Init(runtime.InitOption{Drop: opt.Drop})
	if err == nil {
		ccolor.Infoln("🎉  Migration schema initialized successfully.")
	}
	return err

	/*
	   	if err := initConfigAndDB(); err != nil {
	   		return err
	   	}

	   defer cleanupDB()

	   // Drop existing schema if needed

	   	if opt.Drop {
	   		if err := db.DropSchema(); err != nil {
	   			return fmt.Errorf("failed to drop schema: %v", err)
	   		}
	   	}

	   // Initialize schema if needed

	   	if err := db.InitSchema(); err != nil {
	   		return fmt.Errorf("failed to initialize schema: %v", err)
	   	}

	   ccolor.Infoln("🎉  Migration schema initialized successfully.")
	   return nil
	*/
}
