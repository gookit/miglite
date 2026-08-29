package runtime

import "fmt"

func (r *Runtime) Init(opt InitOption) error {
	if err := r.ensureDB(); err != nil {
		return err
	}
	if opt.Drop {
		if err := r.db.DropSchema(); err != nil {
			return fmt.Errorf("failed to drop schema: %v", err)
		}
	}
	if err := r.db.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize schema: %v", err)
	}
	return nil
}
