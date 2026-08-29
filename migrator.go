package miglite

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/internal/runtime"
	"github.com/gookit/miglite/pkg/command"
)

// Migrator manage the migration
type Migrator struct {
	cfg   *Config
	db    *database.DB
	fsys  fs.FS
	ownDB bool
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

// SetSqlDB sets the database connection
func (m *Migrator) SetSqlDB(db *sql.DB) *Migrator {
	if db == nil {
		m.db = nil
		m.ownDB = false
		return m
	}
	m.db = database.NewWithSqlDB(m.cfg.Database.Driver, db)
	m.ownDB = false
	return m
}

func (m *Migrator) SetFS(fsys fs.FS) *Migrator { m.fsys = fsys; return m }

func (m *Migrator) runtime() *runtime.Runtime {
	r := runtime.NewWithDatabase(m.cfg, m.db, m.ownDB)
	r.SetFS(m.fsys)
	return r
}

func (m *Migrator) Close() error {
	if m.db == nil || !m.ownDB {
		return nil
	}
	err := m.db.Close()
	m.db = nil
	m.ownDB = false
	return err
}

// Init initializes the migration schema
func (m *Migrator) Init(opt command.InitOption) error {
	r := m.runtime()
	defer r.Close()
	return r.Init(runtime.InitOption{Drop: opt.Drop})
}

// Up runs the migration up operation.
func (m *Migrator) Up(opt command.UpOption) error {
	r := m.runtime()
	defer r.Close()
	return r.Up(runtime.UpOption{Yes: opt.Yes, SkipErr: opt.SkipErr, Number: opt.Number, StartTime: opt.StartTime})
}

// Down runs the migration down operation.
func (m *Migrator) Down(opt command.DownOption) error {
	r := m.runtime()
	defer r.Close()
	return r.Down(runtime.DownOption{Number: opt.Number, Yes: opt.Yes})
}

// Skip skips some migration files.
func (m *Migrator) Skip(opt command.SkipOption) error {
	r := m.runtime()
	defer r.Close()
	return r.Skip(runtime.SkipOption{FileNames: opt.FileNames})
}

// Status shows the status of the migrations.
func (m *Migrator) Status(opt command.StatusOption) error {
	r := m.runtime()
	defer r.Close()
	records, err := r.Status(runtime.StatusOption{})
	for _, record := range records {
		fmt.Printf("%s %s\n", record.Status, record.Version)
	}
	return err
}

// Show displays all tables in the database.
func (m *Migrator) Show(opt command.ShowOption) error {
	r := m.runtime()
	defer r.Close()
	result, err := r.Show(runtime.ShowOption{Tables: opt.Tables, Schema: opt.Schema})
	if err == nil {
		fmt.Printf("%v\n", result)
	}
	return err
}

// Exec executes SQL or a SQL file in a transaction.
func (m *Migrator) Exec(opt command.ExecOption) error {
	r := m.runtime()
	defer r.Close()
	return r.Exec(runtime.ExecOption{SQLOrFile: opt.SQLOrFile, Yes: opt.Yes})
}
