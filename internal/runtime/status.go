package runtime

import (
	"errors"
	"github.com/gookit/miglite/internal/migutil"
	"github.com/gookit/miglite/pkg/migration"
)

func (r *Runtime) Status(StatusOption) ([]migration.Record, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	ms, err := migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	if err != nil {
		return nil, err
	}
	statuses, err := migration.GetMigrationsStatus(r.db, ms)
	if err != nil && migutil.IsTableNotExists(r.cfg.Database.Driver, err.Error()) {
		return nil, errors.New("migration table does not exist. please run `miglite init` to create it")
	}
	return statuses, err
}
