package reports

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	envDatabaseURL = "ANALYSIS_DATABASE_URL"
	envHost        = "ANALYSIS_DB_HOST"
	envPort        = "ANALYSIS_DB_PORT"
	envName        = "ANALYSIS_DB_NAME"
	envUser        = "ANALYSIS_DB_USER"
	envPassword    = "ANALYSIS_DB_PASSWORD"
	envPgHost      = "PG_ANALYSIS_HOST"
	envPgPort      = "PG_ANALYSIS_PORT"
	envPgName      = "PG_ANALYSIS_DB"
	envPgUser      = "PG_ANALYSIS_USER"
	envPgPassword  = "PG_ANALYSIS_PASSWORD"
	defaultHost    = "postgres-analysis"
	defaultPort    = "5432"
	defaultSSLMode = "disable"
)

func LoadDSNFromEnv() (string, error) {
	if dsn := strings.TrimSpace(os.Getenv(envDatabaseURL)); dsn != "" {
		return dsn, nil
	}

	host := firstNonEmpty(envHost, envPgHost)
	if host == "" {
		host = defaultHost
	}
	port := firstNonEmpty(envPort, envPgPort)
	if port == "" {
		port = defaultPort
	}
	name := firstNonEmpty(envName, envPgName, "")
	user := firstNonEmpty(envUser, envPgUser, "")
	password := firstNonEmpty(envPassword, envPgPassword, "")

	if name == "" || user == "" {
		return "", errors.New("analysis database name/user are not set")
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		name,
		defaultSSLMode,
	), nil
}

func firstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val
		}
	}
	return ""
}
