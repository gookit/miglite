package runtime

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/gookit/miglite/internal/config"
	"github.com/gookit/miglite/internal/database"
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

func (r *Runtime) SetFS(fsys fs.FS) { r.fsys = fsys }

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
