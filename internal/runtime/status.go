package runtime

import "github.com/gookit/miglite/pkg/migration"

func (r *Runtime) Status(StatusOption) ([]migration.Record, error) {
	if err := r.ensureDB(); err != nil {
		return nil, err
	}
	ms, err := migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	if err != nil {
		return nil, err
	}
	return migration.GetMigrationsStatus(r.db, ms)
}
