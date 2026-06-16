package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/envutil"
	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite/internal/config"
)

func TestLoadWithEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	defaultEnv := filepath.Join(tmpDir, ".env")
	customEnv := filepath.Join(tmpDir, "custom.env")

	assert.NoErr(t, os.WriteFile(defaultEnv, []byte("DATABASE_URL=sqlite://default.db\n"), 0644))
	assert.NoErr(t, os.WriteFile(customEnv, []byte("DATABASE_URL=sqlite://custom.db\nMIGRATIONS_PATH=./custom_migrations\n"), 0644))

	oldWd, err := os.Getwd()
	assert.NoErr(t, err)
	assert.NoErr(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		assert.NoErr(t, os.Chdir(oldWd))
	})

	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = customEnv
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
	})

	cfg, err := config.Load("missing.yaml")
	assert.NoErr(t, err)
	assert.Eq(t, "sqlite", cfg.Database.Driver)
	assert.Eq(t, "custom.db", cfg.Database.DSN)
	assert.Eq(t, "./custom_migrations", cfg.Migrations.Path)
	assert.Eq(t, []string{customEnv}, envutil.LoadedEnvFiles())
}

func TestLoadWithMissingEnvFile(t *testing.T) {
	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = filepath.Join(t.TempDir(), "missing.env")
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
	})

	_, err := config.Load("missing.yaml")
	assert.Err(t, err)
	assert.Contains(t, err.Error(), "missing.env")
}
