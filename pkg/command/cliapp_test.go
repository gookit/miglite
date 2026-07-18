package command

import (
	"testing"

	"github.com/gookit/goutil/cflag/capp"
	"github.com/gookit/goutil/envutil"
	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/miglite/internal/config"
)

func TestNewAppEnvFileFlag(t *testing.T) {
	t.Cleanup(func() {
		envFile = ""
		config.EnvFile = ""
	})

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "long flag", args: []string{"--env-file", "custom.env", "noop"}, want: "custom.env"},
		{name: "alias flag", args: []string{"--efile", "alias.env", "noop"}, want: "alias.env"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile = ""
			config.EnvFile = ""

			app := NewApp("miglite", "test", "test app")
			app.Add(capp.NewCmd("noop", "noop"))
			app.BeforeRun = func(c *capp.Cmd, cmdArgs []string) bool {
				return false
			}

			err := app.RunWithArgs(tt.args)
			assert.NoErr(t, err)
			assert.Eq(t, tt.want, config.EnvFile)
		})
	}
}

func TestBindCommonFlagsEnvFile(t *testing.T) {
	t.Cleanup(func() {
		envFile = ""
		config.EnvFile = ""
	})

	cmd := capp.NewCmd("noop", "noop")
	bindCommonFlags(cmd)

	err := cmd.Parse([]string{"--env-file", "cmd.env"})
	assert.NoErr(t, err)
	syncEnvOptions()
	assert.Eq(t, "cmd.env", config.EnvFile)
}

func TestConfigFlagDefault(t *testing.T) {
	t.Cleanup(func() {
		ConfigFile = ""
	})

	cmd := capp.NewCmd("noop", "noop")
	bindCommonFlags(cmd)

	err := cmd.Parse([]string{})
	assert.NoErr(t, err)
	assert.Eq(t, "", ConfigFile)
}

func TestConfigFlagValue(t *testing.T) {
	t.Cleanup(func() {
		ConfigFile = ""
	})

	cmd := capp.NewCmd("noop", "noop")
	bindCommonFlags(cmd)

	err := cmd.Parse([]string{"--config", "custom.yaml"})
	assert.NoErr(t, err)
	assert.Eq(t, "custom.yaml", ConfigFile)
}

func TestDBFlag(t *testing.T) {
	t.Cleanup(func() { DBName = "" })

	t.Run("common command flag", func(t *testing.T) {
		DBName = ""
		cmd := capp.NewCmd("noop", "noop")
		bindCommonFlags(cmd)
		assert.NoErr(t, cmd.Parse([]string{"--db", "common_db"}))
		assert.Eq(t, "common_db", DBName)
	})

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "root flag", args: []string{"--db", "root_db", "noop"}, want: "root_db"},
		{name: "command flag", args: []string{"noop", "--db", "command_db"}, want: "command_db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DBName = ""
			app := NewApp("miglite", "test", "test app")
			noop := capp.NewCmd("noop", "noop")
			bindCommonFlags(noop)
			app.Add(noop)

			assert.NoErr(t, app.RunWithArgs(tt.args))
			assert.Eq(t, tt.want, DBName)
		})
	}
}

func TestInitLoadConfigDBOverride(t *testing.T) {
	t.Setenv(config.EnvDBURL, "sqlite://old.db")
	t.Setenv(config.EnvDBDSN, "")
	t.Setenv(config.EnvDBDriver, "")
	t.Setenv(config.EnvDBSqlDriver, "")
	t.Setenv(config.EnvPrefixKey, "")
	envutil.StdDotenv().Reset()
	config.EnvPrefix = ""
	config.EnvFile = ""
	cfg = nil
	DBName = "new.db"
	t.Cleanup(func() {
		envutil.StdDotenv().Reset()
		config.EnvPrefix = ""
		config.EnvFile = ""
		cfg = nil
		DBName = ""
	})

	assert.NoErr(t, initLoadConfig())
	assert.Eq(t, "new.db", Cfg().Database.DSN)
}
