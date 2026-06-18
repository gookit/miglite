package command

import (
	"testing"

	"github.com/gookit/goutil/cflag/capp"
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
