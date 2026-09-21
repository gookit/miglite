package runtime

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
	"github.com/gookit/miglite/pkg/migration"
)

type Runtime struct {
	cfg   *config.Config
	db    *database.DB
	fsys  fs.FS
	ownDB bool
}

func New(cfg *config.Config, db *sql.DB) *Runtime {
	r := &Runtime{cfg: cfg}
	if db != nil && cfg != nil {
		r.db = database.NewWithSqlDB(cfg.Database.Driver, db)
	}
	return r
}

func NewWithDatabase(cfg *config.Config, db *database.DB, ownDB bool) *Runtime {
	return &Runtime{cfg: cfg, db: db, ownDB: ownDB}
}

// SetFS sets an explicit migrations filesystem, eg. an embed.FS. Migration
// discovery then resolves cfg.Migrations.Path as an io/fs logical path (slash
// separated) instead of a local directory. nil restores local filesystem reading.
func (r *Runtime) SetFS(fsys fs.FS) { r.fsys = fsys }

// findMigrations discovers migration files from the injected fs.FS when one is
// set, otherwise from the local filesystem.
func (r *Runtime) findMigrations() ([]*migration.Migration, error) {
	if r.fsys != nil {
		return migration.FindMigrationsFS(r.fsys, r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	}
	return migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
}

// parseMigration loads the UP/DOWN sections of a resolved migration from the
// runtime migrations source: from fsys when one is injected, otherwise from the
// local filesystem. Discovered migrations only carry their file name and path.
func (r *Runtime) parseMigration(m *migration.Migration) error {
	if r.fsys != nil {
		return m.LoadFS(r.fsys)
	}
	return m.Parse()
}

// migrationsFrom loads the named migration files from the same source as
// findMigrations, used by the skip operation.
func (r *Runtime) migrationsFrom(files []string) ([]*migration.Migration, error) {
	if r.fsys != nil {
		return migration.MigrationsFromFS(r.fsys, r.cfg.Migrations.Path, files)
	}
	return migration.MigrationsFrom(r.cfg.Migrations.Path, files)
}

func (r *Runtime) Close() error {
	if r.db == nil || !r.ownDB {
		return nil
	}
	err := r.db.Close()
	r.db = nil
	r.ownDB = false
	return err
}

func (r *Runtime) ensureDB() error {
	if r.cfg == nil {
		return fmt.Errorf("runtime config is nil")
	}
	if r.db != nil {
		return nil
	}
	db, err := database.NewDB(r.cfg.Database.Driver, r.cfg.Database.SqlDriver, r.cfg.Database.DSN)
	if err != nil {
		return err
	}
	r.db, r.ownDB = db, true
	return nil
}
