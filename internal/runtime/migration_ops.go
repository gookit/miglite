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
	return r.UpWithHooks(opt, MigrationHooks{})
}

func (r *Runtime) UpWithHooks(opt UpOption, hooks MigrationHooks) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	if err := r.db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}
	ms, err := migration.FindMigrations(r.cfg.Migrations.Path, r.cfg.Migrations.Recursive)
	if err != nil {
		return err
	}
	e := migration.NewExecutor(r.db, r.cfg.Verbose)
	n := 0
	for i, m := range ms {
		applied, status, err := migration.IsApplied(r.db, m.FileName)
		if err != nil {
			return err
		}
		if applied || status == migration.StatusSkip {
			if hooks.Skip != nil {
				hooks.Skip(i, len(ms), m, status)
			}
			continue
		}
		if hooks.Before != nil {
			if err := hooks.Before(i, len(ms), m); err != nil {
				return err
			}
		}
		if err = m.Parse(); err != nil {
			return err
		}
		if err = e.ExecuteUp(m); err != nil {
			if hooks.Error != nil {
				hooks.Error(i, len(ms), m, err)
			}
			if opt.SkipErr {
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %v", m.FileName, err)
		}
		n++
		if hooks.After != nil {
			hooks.After(i, len(ms), m)
		}
		if opt.Number > 0 && n >= opt.Number {
			break
		}
	}
	return nil
}

func (r *Runtime) Down(opt DownOption) error {
	return r.DownWithHooks(opt, MigrationHooks{})
}

func (r *Runtime) DownWithHooks(opt DownOption, hooks MigrationHooks) error {
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
	for i, rec := range recs {
		for _, m := range ms {
			if m.Version == rec.Version {
				if hooks.Before != nil {
					if err := hooks.Before(i, len(recs), m); err != nil {
						return err
					}
				}
				if err = m.Parse(); err != nil {
					return err
				}
				if m.DownSection != "" {
					if err = e.ExecuteDown(m); err != nil {
						if hooks.Error != nil {
							hooks.Error(i, len(recs), m, err)
						}
						return err
					}
				}
				if hooks.After != nil {
					hooks.After(i, len(recs), m)
				}
				break
			}
		}
	}
	return nil
}

func (r *Runtime) Skip(opt SkipOption) error {
	return r.SkipWithHooks(opt, MigrationHooks{})
}

func (r *Runtime) SkipWithHooks(opt SkipOption, hooks MigrationHooks) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	ms, err := migration.MigrationsFrom(r.cfg.Migrations.Path, opt.FileNames)
	if err != nil {
		return err
	}
	for i, m := range ms {
		if hooks.Before != nil {
			if err := hooks.Before(i, len(ms), m); err != nil {
				return err
			}
		}
		if err = migration.SaveRecord(r.db, m.Version, migration.StatusSkip, nil); err != nil {
			if hooks.Error != nil {
				hooks.Error(i, len(ms), m, err)
			}
			return err
		}
		if hooks.After != nil {
			hooks.After(i, len(ms), m)
		}
	}
	return nil
}
