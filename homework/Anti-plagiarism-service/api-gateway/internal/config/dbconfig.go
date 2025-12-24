package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	envGatewayDatabaseURL = "GATEWAY_DATABASE_URL"
	envGatewayHost        = "GATEWAY_DB_HOST"
	envGatewayPort        = "GATEWAY_DB_PORT"
	envGatewayName        = "GATEWAY_DB_NAME"
	envGatewayUser        = "GATEWAY_DB_USER"
	envGatewayPassword    = "GATEWAY_DB_PASSWORD"
	envPgGatewayHost      = "PG_GATEWAY_HOST"
	envPgGatewayPort      = "PG_GATEWAY_PORT"
	envPgGatewayName      = "PG_GATEWAY_DB"
	envPgGatewayUser      = "PG_GATEWAY_USER"
	envPgGatewayPassword  = "PG_GATEWAY_PASSWORD"
	defaultGatewayHost    = "postgres-gateway"
	defaultGatewayPort    = "5432"
	defaultGatewaySSLMode = "disable"
)

func LoadGatewayDSNFromEnv() (string, error) {
	if dsn := strings.TrimSpace(os.Getenv(envGatewayDatabaseURL)); dsn != "" {
		return dsn, nil
	}

	host := firstGatewayNonEmpty(envGatewayHost, envPgGatewayHost)
	if host == "" {
		host = defaultGatewayHost
	}
	port := firstGatewayNonEmpty(envGatewayPort, envPgGatewayPort)
	if port == "" {
		port = defaultGatewayPort
	}
	name := firstGatewayNonEmpty(envGatewayName, envPgGatewayName)
	user := firstGatewayNonEmpty(envGatewayUser, envPgGatewayUser)
	password := firstGatewayNonEmpty(envGatewayPassword, envPgGatewayPassword)

	if name == "" || user == "" {
		return "", errors.New("gateway database name/user are not set")
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		name,
		defaultGatewaySSLMode,
	), nil
}

func firstGatewayNonEmpty(keys ...string) string {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val
		}
	}
	return ""
}
