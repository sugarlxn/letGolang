package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// SetupTestDB 创建测试数据库并自动运行所有 migrations
// 使用方式：
//
//	db := testutil.SetupTestDB(t)
//	defer db.Close()
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// 创建临时数据库文件
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// 自动运行所有 migration 文件
	if err := runMigrations(db, t); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// runMigrations 自动读取并执行 migrations 目录中的所有 SQL 文件
func runMigrations(db *sql.DB, t *testing.T) error {
	t.Helper()

	// 获取项目根目录的 migrations 文件夹路径
	migrationsDir := getMigrationsDir()

	// 读取所有 .sql 文件
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// 按文件名排序（确保按顺序执行）
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// 依次执行每个 migration 文件
	for _, filename := range sqlFiles {
		filePath := filepath.Join(migrationsDir, filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// 执行 SQL（跳过注释和 DOWN 部分）
		sql := cleanSQL(string(content))

		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		t.Logf("✓ Applied migration: %s", filename)
	}

	return nil
}

// getMigrationsDir 获取 migrations 目录的绝对路径
func getMigrationsDir() string {
	// 从当前文件位置向上查找 migrations 目录
	// 适配不同的测试执行路径
	currentDir, _ := os.Getwd()

	// 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join(currentDir, "migrations"),
		filepath.Join(currentDir, "..", "migrations"),
		filepath.Join(currentDir, "webserver", "migrations"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 默认路径
	return "./migrations"
}

// cleanSQL 清理 SQL 内容：
// - 移除单行注释 (-- ...)
// - 移除 DOWN 部分
// - 保留实际的 SQL 语句
func cleanSQL(content string) string {
	lines := strings.Split(content, "\n")
	var cleaned []string
	inDownSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测 DOWN 部分开始
		if strings.Contains(trimmed, "-- DOWN:") ||
			strings.Contains(trimmed, "========================================") && inDownSection {
			inDownSection = true
			continue
		}

		// 跳过注释行和空行
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}

		// 如果在 DOWN 部分，跳过
		if inDownSection {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// SeedTestData 为测试提供一些种子数据（可选）
func SeedTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	// 示例：插入测试用户
	_, err := db.Exec(`
		INSERT INTO users (username, password, phone, email) VALUES 
		('testuser1', '$2a$10$hash1', '1234567890', 'test1@example.com'),
		('testuser2', '$2a$10$hash2', '0987654321', 'test2@example.com')
	`)
	if err != nil {
		t.Fatalf("failed to seed test data: %v", err)
	}

	t.Log("✓ Test data seeded")
}
