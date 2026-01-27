package miglite

import (
	"database/sql"

	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/config"
	"github.com/gookit/miglite/pkg/database"
)

// ConfigFn is a function type for updating the configuration
type ConfigFn func(c *config.Config)

// Migrator manage the migration
type Migrator struct {
	cfg *config.Config
	// TODO add migrations by go code
	//
	// Example:
	// 	mig.Add("2026...-add_user_table", `UP sql`, `DOWN sql`, options)
	//
	// migrations []*migration.Migration
}

// New creates a new Migrator instance, with optional configuration functions
//
//  - configFile: if not exist, will skip load it.
func New(configFile string, fns ...ConfigFn) (*Migrator, error) {
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, err
	}

	for _, fn := range fns {
		fn(cfg)
	}
	return NewWithConfig(cfg), nil
}

// NewWithConfig creates a new Migrator instance with a pre-configured Config
func NewWithConfig(cfg *config.Config) *Migrator {
	command.SetCfg(cfg)
	return &Migrator{cfg: cfg}
}

// SetSqlDB sets the database connection
func (m *Migrator) SetSqlDB(db *sql.DB) {
	dbCfg := m.cfg.Database
	command.SetDB(database.NewWithSqlDB(dbCfg.Driver, db))
}

// Init initializes the migration schema
func (m *Migrator) Init(opt command.InitOption) error {
	return command.HandleInit(opt)
}

// Up runs the migration up operation.
func (m *Migrator) Up(opt command.UpOption) error {
	return command.HandleUp(opt)
}

// Down runs the migration down operation.
func (m *Migrator) Down(opt command.DownOption) error {
	return command.HandleDown(opt)
}

// Skip skips some migration files.
func (m *Migrator) Skip(opt command.SkipOption) error {
	return command.HandleSkip(opt)
}

// Status shows the status of the migrations.
func (m *Migrator) Status(opt command.StatusOption) error {
	return command.HandleStatus(opt)
}

// Show displays all tables in the database.
func (m *Migrator) Show(opt command.ShowOption) error {
	return command.HandleShow(opt)
}
