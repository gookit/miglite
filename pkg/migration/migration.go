package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/goutil/fsutil"
)

// Migration represents a single migration file
type Migration struct {
	FileName string // filename as version
	FilePath string // full path
	// Contents of migration file
	Contents string
	// time from filename
	Timestamp time.Time
	// Version same as filename
	Version string
	// UpSection UP section contents
	UpSection   string
	DownSection string
	// Options TODO options for current migration
}

// ParseFile parses a migration file to extract UP and DOWN sections
func ParseFile(filePath string) (*Migration, error) {
	migFile, err := NewMigration(filePath)
	if err != nil {
		return nil, err
	}

	if err1 := migFile.Parse(); err1 != nil {
		return nil, err1
	}
	return migFile, nil
}

// MigrationsFrom creates migrations from a list of file names
func MigrationsFrom(migPath string, files []string) ([]*Migration, error) {
	migrations := make([]*Migration, 0, len(files))
	var migPaths []string
	if strings.Contains(migPath, ",") {
		migPaths = strings.Split(migPath, ",")
	} else {
		migPaths = []string{migPath}
	}

	for _, file := range files {
		var filePath string
		var fileExists bool
		// check and auto append .sql
		if !strings.HasSuffix(file, ".sql") {
			file = file + ".sql"
		}

		// is multiple paths
		if len(migPaths) > 0 {
			for _, dirPath := range migPaths {
				if fsutil.IsFile(dirPath + "/" + file) {
					filePath = dirPath + "/" + file
					fileExists = true
					break
				}
			}
		} else {
			filePath = migPath + "/" + file
			fileExists = fsutil.IsFile(filePath)
		}

		if !fileExists {
			return nil, fmt.Errorf("migration file not exists: %s", file)
		}

		mig, err := NewMigration(filePath)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, mig)
	}

	return migrations, nil
}

// NewMigration creates a new Migration instance from a file path
func NewMigration(filePath string) (*Migration, error) {
	// Extract timestamp from filename
	fileName := filepath.Base(filePath)
	fi, err := parseFilename(fileName)
	if err != nil {
		return nil, err
	}

	return &Migration{
		FileName:  fileName,
		FilePath:  filePath,
		Timestamp: fi.Time,
		Version:   fileName,
	}, nil
}

// Parse reads migration file and parse it contents.
func (m *Migration) Parse() error {
	contents, err := os.ReadFile(m.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %s", err)
	}

	m.Contents = string(contents)
	return m.ParseContents()
}

// ParseContents parses migration file contents, extracting UP and DOWN sections
func (m *Migration) ParseContents() error {
	if m.Contents == "" {
		return fmt.Errorf("migration file contents is empty. file: %s", m.FilePath)
	}

	// 使用按行解析处理
	lines := strings.Split(m.Contents, "\n")
	var upLines, downLines []string
	// "" for none, "up" for up section, "down" for down section
	currentSection := ""

	// TODO 后续支持在前几行设置选项。格式：-- Migrate-option:OPTION=VALUE,OPTION1=VALUE1,
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 跳过空行
		if trimmed == "" {
			continue
		}

		// TODO 后续支持 UP, DOWN 后面跟自定义设置: -- Migrate:UP(option=value,)
		if strings.HasPrefix(trimmed, MarkUp) {
			currentSection = "up"
			continue
		} else if strings.HasPrefix(trimmed, MarkDown) {
			currentSection = "down"
			continue
		}

		// 跳过不需要的注释行
		if strings.HasPrefix(trimmed, "-- ") {
			continue
		}

		// 根据当前部分添加行内容
		switch currentSection {
		case "up":
			upLines = append(upLines, line)
		case "down":
			downLines = append(downLines, line)
		}
	}

	// 设置解析结果
	m.UpSection = strings.Join(upLines, "\n")
	if len(downLines) > 0 {
		m.DownSection = strings.Join(downLines, "\n")
	}

	// 验证必须包含 UP 部分
	if m.UpSection == "" {
		return fmt.Errorf("migration file %s does not contain valid SQL in '-- Migrate:UP' section", m.FilePath)
	}
	return nil
}

// ResetContents 重置迁移文件内容字段
func (m *Migration) ResetContents() {
	m.Contents = ""
	m.UpSection = ""
	m.DownSection = ""
}

// IsBefore 判断当前迁移文件是否早于指定迁移文件
func (m *Migration) IsBefore(other *Migration) bool {
	return m.Timestamp.Before(other.Timestamp)
}
