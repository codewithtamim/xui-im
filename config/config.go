// Package config provides configuration management utilities for the xui-im panel,
// including version information, logging levels, database paths, and environment variable handling.
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

// LogLevel represents the logging level for the application.
type LogLevel string

// Logging level constants
const (
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Notice  LogLevel = "notice"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

// GetVersion returns the version string of the xui-im application.
func GetVersion() string {
	return strings.TrimSpace(version)
}

// GetName returns the name of the xui-im application.
func GetName() string {
	return strings.TrimSpace(name)
}

// GetLogLevel returns the current logging level based on environment variables or defaults to Info.
func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("XUI_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

// IsDebug returns true if debug mode is enabled via the XUI_DEBUG environment variable.
func IsDebug() bool {
	return os.Getenv("XUI_DEBUG") == "true"
}

// GetBinFolderPath returns the path to the binary folder, defaulting to "bin" if not set via XUI_BIN_FOLDER.
func GetBinFolderPath() string {
	binFolderPath := os.Getenv("XUI_BIN_FOLDER")
	if binFolderPath == "" {
		binFolderPath = "bin"
	}
	return binFolderPath
}

func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exeDir := filepath.Dir(exePath)
	exeDirLower := strings.ToLower(filepath.ToSlash(exeDir))
	if strings.Contains(exeDirLower, "/appdata/local/temp/") || strings.Contains(exeDirLower, "/go-build") {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return exeDir
}

// GetDBDSN returns the PostgreSQL connection string.
func GetDBDSN() string {
	return GetPostgresDSN()
}

// GetPostgresDSN builds a PostgreSQL connection string from environment variables.
// Supported env vars: XUI_DB_HOST, XUI_DB_PORT, XUI_DB_USER, XUI_DB_PASSWORD,
// XUI_DB_NAME, XUI_DB_SSL_MODE.
func GetPostgresDSN() string {
	host := os.Getenv("XUI_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("XUI_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("XUI_DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("XUI_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("XUI_DB_NAME")
	if dbname == "" {
		dbname = "xui"
	}
	sslmode := os.Getenv("XUI_DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// GetPostgresEnv returns PostgreSQL connection parameters as individual values
// for use with command-line tools like pg_dump.
func GetPostgresEnv() (host, port, user, password, dbname string) {
	host = os.Getenv("XUI_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port = os.Getenv("XUI_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user = os.Getenv("XUI_DB_USER")
	if user == "" {
		user = "postgres"
	}
	password = os.Getenv("XUI_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname = os.Getenv("XUI_DB_NAME")
	if dbname == "" {
		dbname = "xui"
	}
	return
}

// GetLogFolder returns the path to the log folder based on environment variables or platform defaults.
func GetLogFolder() string {
	logFolderPath := os.Getenv("XUI_LOG_FOLDER")
	if logFolderPath != "" {
		return logFolderPath
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(".", "log")
	}
	return "/var/log/x-ui"
}


