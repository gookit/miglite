package runtime

import (
	"fmt"
	"github.com/gookit/miglite/pkg/migration"
)

type UpOption struct {
	Yes, SkipErr bool
	Number       int
	StartTime    string
}
type DownOption struct {
	Number int
	Yes    bool
}
type SkipOption struct{ FileNames []string }

func (r *Runtime) Up(opt UpOption) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	ms, err := migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	if err != nil {
		return err
	}
	e := migration.NewExecutor(r.db, r.cfg.Verbose)
	n := 0
	for _, m := range ms {
		applied, status, err := migration.IsApplied(r.db, m.FileName)
		if err != nil {
			return err
		}
		if applied || status == migration.StatusSkip {
			continue
		}
		if err = m.Parse(); err != nil {
			return err
		}
		if err = e.ExecuteUp(m); err != nil {
			if opt.SkipErr {
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %v", m.FileName, err)
		}
		n++
		if opt.Number > 0 && n >= opt.Number {
			break
		}
	}
	return nil
}

func (r *Runtime) Down(opt DownOption) error {
	if opt.Number <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}
	if err := r.ensureDB(); err != nil {
		return err
	}
	recs, err := migration.GetAppliedSortedByVersion(r.db, opt.Number)
	if err != nil {
		return err
	}
	ms, err := migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	if err != nil {
		return err
	}
	e := migration.NewExecutor(r.db, r.cfg.Verbose)
	for _, rec := range recs {
		for _, m := range ms {
			if m.Version == rec.Version {
				if err = m.Parse(); err != nil {
					return err
				}
				if m.DownSection != "" {
					if err = e.ExecuteDown(m); err != nil {
						return err
					}
				}
				break
			}
		}
	}
	return nil
}

func (r *Runtime) Skip(opt SkipOption) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	ms, err := migration.MigrationsFrom(r.cfg.Migrations.Path, opt.FileNames)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if err = migration.SaveRecord(r.db, m.Version, migration.StatusSkip, nil); err != nil {
			return err
		}
	}
	return nil
}
