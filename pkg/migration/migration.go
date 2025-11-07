package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// defines the regex pattern for extracting the date prefix from a filename
//
// format: YYYYMMDD-HHMMSS-{name}.sql
var regexFilename = regexp.MustCompile(`^(\d{8}-\d{6})-([\w-]+)\.sql$`)

// Migration represents a single migration file
type Migration struct {
	FileName string
	FilePath string
	// time from filename
	Timestamp time.Time
	// Version same as filename
	Version string
	// Contents of migration file
	Contents    string
	UpSection   string
	DownSection string
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
		Version: fileName,
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
		if strings.HasSuffix(trimmed, MarkUp) {
			currentSection = "up"
			continue
		} else if strings.HasSuffix(trimmed, MarkDown) {
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
		return fmt.Errorf("migration file %s does not contain valid '-- Migrate:UP' section", m.FilePath)
	}
	return nil
}

// ResetContents 重置迁移文件内容字段
func (m *Migration) ResetContents() {
	m.Contents = ""
	m.UpSection = ""
	m.DownSection = ""
}
