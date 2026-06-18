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

func TestLoadExpandsEnvVarsInConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "config.env")
	configFile := filepath.Join(tmpDir, "miglite.yaml")

	assert.NoErr(t, os.WriteFile(envFile, []byte(`
PG_HOST=127.0.0.1
PG_USER=mig_user
PG_DB_NAME=mig_test
APP_MODULE=billing
`), 0644))
	assert.NoErr(t, os.WriteFile(configFile, []byte(`
database:
  driver: postgres
  host: ${PG_HOST}
  port: 15432
  user: ${PG_USER}
  password: ${PG_PASSWORD | fallback_secret}
  dbname: ${PG_DB_NAME}
  ssl_mode: disable
migrations:
  path: ./migrations/${APP_MODULE}
`), 0644))

	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = envFile
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
	})

	cfg, err := config.Load(configFile)
	assert.NoErr(t, err)
	assert.Eq(t, configFile, cfg.ConfigFile)
	assert.Eq(t, "postgres", cfg.Database.Driver)
	assert.Eq(t, "127.0.0.1", cfg.Database.Host)
	assert.Eq(t, 15432, cfg.Database.Port)
	assert.Eq(t, "mig_user", cfg.Database.User)
	assert.Eq(t, "fallback_secret", cfg.Database.Password)
	assert.Eq(t, "mig_test", cfg.Database.DBName)
	assert.Eq(t, "disable", cfg.Database.SSLMode)
	assert.Eq(t, "./migrations/billing", cfg.Migrations.Path)
}

func TestLoadReturnsConfigEnvParseError(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "miglite.yaml")
	assert.NoErr(t, os.WriteFile(configFile, []byte(`
database:
  driver: sqlite
  dsn: ${MISSING_SQLITE_DSN | ?missing sqlite dsn}
`), 0644))

	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = ""
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
	})

	_, err := config.Load(configFile)
	assert.Err(t, err)
	assert.ErrMsg(t, err, "missing sqlite dsn")
}

func TestLoadUsesDefaultConfigFiles(t *testing.T) {
	t.Run("loads miglite yaml first", func(t *testing.T) {
		clearConfigEnv(t)
		tmpDir := t.TempDir()
		assert.NoErr(t, os.WriteFile(filepath.Join(tmpDir, "miglite.yaml"), []byte(`
database:
  driver: sqlite
  dsn: default.db
`), 0644))
		assert.NoErr(t, os.WriteFile(filepath.Join(tmpDir, "miglite.local.yaml"), []byte(`
database:
  driver: sqlite
  dsn: local.db
`), 0644))

		oldWd, err := os.Getwd()
		assert.NoErr(t, err)
		assert.NoErr(t, os.Chdir(tmpDir))
		t.Cleanup(func() {
			assert.NoErr(t, os.Chdir(oldWd))
		})

		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
		t.Cleanup(func() {
			envutil.StdDotenv().Reset()
			config.EnvPrefix = ""
			config.EnvFile = ""
		})

		cfg, err := config.Load("")
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, "./miglite.yaml", cfg.ConfigFile)
		assert.Eq(t, "default.db", cfg.Database.DSN)
	})

	t.Run("loads local yaml when default is missing", func(t *testing.T) {
		clearConfigEnv(t)
		tmpDir := t.TempDir()
		assert.NoErr(t, os.WriteFile(filepath.Join(tmpDir, "miglite.local.yaml"), []byte(`
database:
  driver: sqlite
  dsn: local.db
`), 0644))

		oldWd, err := os.Getwd()
		assert.NoErr(t, err)
		assert.NoErr(t, os.Chdir(tmpDir))
		t.Cleanup(func() {
			assert.NoErr(t, os.Chdir(oldWd))
		})

		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
		t.Cleanup(func() {
			envutil.StdDotenv().Reset()
			config.EnvPrefix = ""
			config.EnvFile = ""
		})

		cfg, err := config.Load("")
		assert.Require(t, assert.NoErr(t, err))
		assert.Eq(t, "./miglite.local.yaml", cfg.ConfigFile)
		assert.Eq(t, "local.db", cfg.Database.DSN)
	})
}

func TestLoadExplicitConfigFileDoesNotFallback(t *testing.T) {
	clearConfigEnv(t)
	tmpDir := t.TempDir()
	assert.NoErr(t, os.WriteFile(filepath.Join(tmpDir, "miglite.yaml"), []byte(`
database:
  driver: sqlite
  dsn: default.db
`), 0644))

	oldWd, err := os.Getwd()
	assert.NoErr(t, err)
	assert.NoErr(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		assert.NoErr(t, os.Chdir(oldWd))
	})

	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = ""
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
	})

	_, err = config.Load("missing.yaml")
	assert.Err(t, err)
	assert.ErrMsg(t, err, "database driver is required")
}

func clearConfigEnv(t *testing.T) {
	t.Setenv(config.EnvDBDSN, "")
	t.Setenv(config.EnvDBDriver, "")
	t.Setenv(config.EnvDBSqlDriver, "")
	t.Setenv(config.EnvDBURL, "")
	t.Setenv(config.EnvMigrationsPath, "")
	t.Setenv(config.EnvPrefixKey, "")
}
