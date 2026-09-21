package miglite

import (
	"database/sql"
	"io/fs"

	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/command"
)

// Migrator manage the migration
type Migrator struct {
	cfg  *Config
	db   *database.DB
	fsys fs.FS
	// TODO add migrations by go code
	//
	// Example:
	// 	mig.Add("2026...-add_user_table", `UP sql`, `DOWN sql`, options)
	//
	// migrations []*migration.Migration
}

// NewAuto creates a new Migrator instance with autoload default config
func NewAuto(fns ...ConfigFn) (*Migrator, error) {
	return New("", fns...)
}

// New creates a new Migrator instance, with optional configuration functions
//
//   - configFile: if not exist, will skip load it.
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
func NewWithConfig(cfg *Config) *Migrator {
	return &Migrator{cfg: cfg}
}

// NewWithConfigAndFS creates a new Migrator instance with a pre-configured
// Config and an explicit migrations filesystem, eg. an embed.FS. It equals
// NewWithConfig(cfg).SetFS(fsys).
func NewWithConfigAndFS(cfg *Config, fsys fs.FS) *Migrator {
	return NewWithConfig(cfg).SetFS(fsys)
}

// SetSqlDB sets the database connection. The connection stays owned by the
// caller: Migrator never closes it.
func (m *Migrator) SetSqlDB(db *sql.DB) *Migrator {
	if db == nil {
		m.db = nil
		return m
	}
	m.db = database.NewWithSqlDB(m.cfg.Database.Driver, db)
	return m
}

// SetFS sets the migrations filesystem, so the SQL files can be embedded with
// Go's embed.FS. cfg.Migrations.Path is then resolved as an io/fs logical path
// (slash separated, eg: "migrations"). SetFS(nil) restores local filesystem
// reading.
func (m *Migrator) SetFS(fsys fs.FS) *Migrator { m.fsys = fsys; return m }

func (m *Migrator) runtime() *runtime.Runtime {
	// injected connections stay owned by the caller; connections opened from the
	// config are owned and closed by this per-call runtime
	r := runtime.NewWithDatabase(m.cfg, m.db, false)
	r.SetFS(m.fsys)
	return r
}

// Init initializes the migration schema
func (m *Migrator) Init(opt command.InitOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunInit(r, opt)
}

// Up runs the migration up operation.
//
// NOTE: UpOption.Yes only affects the CLI; library calls never ask for
// confirmation. With UpOption.SkipErr failing files are skipped and the run
// continues, but an error listing them is still returned.
func (m *Migrator) Up(opt command.UpOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunUp(r, opt, false)
}

// Down runs the migration down operation.
//
// NOTE: DownOption.Yes only affects the CLI; library calls never ask for
// confirmation.
func (m *Migrator) Down(opt command.DownOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunDown(r, opt, false)
}

// Skip skips some migration files.
func (m *Migrator) Skip(opt command.SkipOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunSkip(r, opt)
}

// Status shows the status of the migrations.
func (m *Migrator) Status(opt command.StatusOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunStatus(r)
}

// Show displays all tables in the database.
func (m *Migrator) Show(opt command.ShowOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunShow(r, opt)
}

// Exec executes SQL or a SQL file in a transaction.
//
// NOTE: ExecOption.Yes only affects the CLI; library calls never ask for
// confirmation.
func (m *Migrator) Exec(opt command.ExecOption) error {
	r := m.runtime()
	defer r.Close()
	return command.RunExec(r, opt, false)
}
