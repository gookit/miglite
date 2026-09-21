package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gookit/miglite/pkg/migration"
)

// StatusEmptyDown marks a migration file without a DOWN section, so
// callers can tell an empty rollback apart from a real one.
const StatusEmptyDown = "empty_down"

type UpOption struct {
	SkipErr   bool
	Number    int
	StartTime string
}
type DownOption struct {
	Number int
}
type SkipOption struct{ FileNames []string }

func (r *Runtime) Up(opt UpOption) error {
	return r.UpWithHooks(opt, MigrationHooks{})
}

// UpWithHooks runs pending migrations.
//
// Hooks contract: Start always reports the total found (0 included); Complete
// fires only when the run finished on its own (not aborted by a hard error or
// by a cancellation). With SkipErr, failed files are skipped, the run continues,
// and an error listing them is still returned so callers cannot report success.
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
	result := MigrationResult{Total: len(ms)}
	if hooks.Start != nil {
		hooks.Start(len(ms))
	}
	if len(ms) == 0 {
		return nil
	}

	completed := false
	defer func() {
		if completed && hooks.Complete != nil {
			hooks.Complete(result)
		}
	}()

	e := migration.NewExecutor(r.db, r.cfg.Verbose)
	n := 0
	var failed []string
	for i, m := range ms {
		applied, status, err := migration.IsApplied(r.db, m.FileName)
		if err != nil {
			return err
		}
		if applied || status == migration.StatusSkip {
			result.Skipped++
			if hooks.Skip != nil {
				hooks.Skip(i, len(ms), m, status)
			}
			continue
		}
		if hooks.Before != nil {
			if err := hooks.Before(i, len(ms), m); err != nil {
				if errors.Is(err, ErrCancelled) {
					return nil
				}
				return err
			}
		}
		if err = m.Parse(); err != nil {
			result.Failed++
			return fmt.Errorf("failed to parse migration %s: %v", m.FileName, err)
		}
		if err = e.ExecuteUp(m); err != nil {
			result.Failed++
			failed = append(failed, m.FileName)
			if hooks.Error != nil {
				hooks.Error(i, len(ms), m, err)
			}
			if opt.SkipErr {
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %v", m.FileName, err)
		}
		n++
		result.Applied++
		m.ResetContents()
		if hooks.After != nil {
			hooks.After(i, len(ms), m)
		}
		if opt.Number > 0 && n >= opt.Number {
			break
		}
	}

	completed = true
	if len(failed) > 0 {
		return fmt.Errorf("%d migration(s) failed: %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}

func (r *Runtime) Down(opt DownOption) error {
	return r.DownWithHooks(opt, MigrationHooks{})
}

// DownWithHooks rolls back applied migrations.
//
// A migration file without a DOWN section is reported through the Skip hook with
// StatusEmptyDown; the After hook runs only for a real rollback, so a
// skipped file is never reported as rolled back.
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
	if hooks.Start != nil {
		hooks.Start(len(recs))
	}

	e := migration.NewExecutor(r.db, r.cfg.Verbose)
	for i, rec := range recs {
		found := false
		cancelled := false
		for _, m := range ms {
			if m.Version != rec.Version {
				continue
			}
			found = true
			m.AppliedAt = rec.AppliedAt
			if hooks.Before != nil {
				if err := hooks.Before(i, len(recs), m); err != nil {
					if errors.Is(err, ErrCancelled) {
						cancelled = true
						continue
					}
					return err
				}
			}
			if err = m.Parse(); err != nil {
				return fmt.Errorf("failed to parse migration %s: %v", m.FileName, err)
			}
			if m.DownSection == "" {
				if hooks.Skip != nil {
					hooks.Skip(i, len(recs), m, StatusEmptyDown)
				}
				break
			}
			if err = e.ExecuteDown(m); err != nil {
				if hooks.Error != nil {
					hooks.Error(i, len(recs), m, err)
				}
				return fmt.Errorf("failed to roll back migration %s: %v", m.FileName, err)
			}
			if hooks.After != nil {
				hooks.After(i, len(recs), m)
			}
			break
		}
		if cancelled {
			continue
		}
		if !found {
			return fmt.Errorf("migration file not found for version: %s", rec.Version)
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
	if hooks.Start != nil {
		hooks.Start(len(ms))
	}
	for i, m := range ms {
		applied, status, err := migration.IsApplied(r.db, m.Version)
		if err != nil {
			return err
		}
		if applied && status == migration.StatusUp {
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
