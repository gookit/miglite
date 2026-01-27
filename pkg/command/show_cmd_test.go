package command

import (
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestShowCommand(t *testing.T) {
	// 测试参数验证
	t.Run("Validation", func(t *testing.T) {
		// 测试没有提供任何选项的情况
		showOpt := ShowOption{}
		err := HandleShow(showOpt)
		assert.Err(t, err)
		assert.Contains(t, err.Error(), "either --tables or --schema must be provided")

		// 测试同时提供两个选项的情况
		showOpt = ShowOption{
			Tables: true,
			Schema: "test_table",
		}
		err = HandleShow(showOpt)
		assert.Err(t, err)
		assert.Contains(t, err.Error(), "--tables and --schema cannot be used together")
	})
}
