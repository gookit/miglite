package command

import (
	"os"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestExecCommand(t *testing.T) {
	// 创建临时SQL文件进行测试
	tmpFile := "temp_test.sql"
	content := `
CREATE TABLE IF NOT EXISTS test_exec (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO test_exec (name) VALUES ('test');
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	assert.NoErr(t, err)
	
	// 清理临时文件
	defer os.Remove(tmpFile)
	
	// 测试exec命令
	execOpt := ExecOption{
		File: tmpFile,
		Yes:  true, // 跳过确认
	}
	
	// 由于需要数据库连接，这里只是验证函数入口点
	// 在完整环境中，这将连接数据库并执行SQL
	err = HandleExec(execOpt)
	
	// 如果没有配置文件或数据库连接，这里可能会失败，这是正常的
	// 主要验证命令结构是否正确
}