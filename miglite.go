package miglite

import (
	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/config"
)

// ConfigFn is a function type for updating the configuration
type ConfigFn func(c *config.Config)

// MigLite manage the migration
type MigLite struct {
	cfg *config.Config
	// TODO add migrations by go code
	//
	// Example:
	// 	mig.Add("2026...-add_user_table", `UP sql`, `DOWN sql`, options)
	//
	// migrations []*migration.Migration
}

// New creates a new MigLite instance, with optional configuration functions
//
//  - configFile: if not exist, will skip load it.
func New(configFile string, fns ...ConfigFn) (*MigLite, error) {
	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, err
	}

	for _, fn := range fns {
		fn(cfg)
	}
	return NewWithConfig(cfg), nil
}

// NewWithConfig creates a new MigLite instance with a pre-configured Config
func NewWithConfig(cfg *config.Config) *MigLite {
	command.SetCfg(cfg)

	return &MigLite{cfg: cfg}
}

// Init initializes the migration schema
func (m *MigLite) Init(opt command.InitOption) error {
	return command.HandleInit(opt)
}

// Up runs the migration up operation.
func (m *MigLite) Up(opt command.UpOption) error {
	return command.HandleUp(opt)
}

// Down runs the migration down operation.
func (m *MigLite) Down(opt command.DownOption) error {
	return command.HandleDown(opt)
}

// Skip skips some migration files.
func (m *MigLite) Skip(opt command.SkipOption) error {
	return command.HandleSkip(opt)
}

// Status shows the status of the migrations.
func (m *MigLite) Status(opt command.StatusOption) error {
	return command.HandleStatus(opt)
}

// Show displays all tables in the database.
func (m *MigLite) Show(opt command.ShowOption) error {
	return command.HandleShow(opt)
}
