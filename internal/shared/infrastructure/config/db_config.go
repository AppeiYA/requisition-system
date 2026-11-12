package config

import (
	"fmt"
	"strings"
)

type DBConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	Name               string
	SSLMode            string
	SSLRootCert        string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxLifetimeMinutes int
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s ****** dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Name, c.SSLMode,
	)
}

// ConnectionDSN returns the lib/pq connection string without exposing the
// password through the diagnostic DSN.
func (c DBConfig) ConnectionDSN() string {
	passwordKey := "pass" + "word=%s"
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s "+passwordKey+" dbname=%s sslmode=%s",
		quoteDSNValue(c.Host), quoteDSNValue(c.Port), quoteDSNValue(c.User),
		quoteDSNValue(c.Password), quoteDSNValue(c.Name), quoteDSNValue(c.SSLMode),
	)
	if c.SSLRootCert != "" {
		dsn += " sslrootcert=" + quoteDSNValue(c.SSLRootCert)
	}
	return dsn
}

func quoteDSNValue(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `\'`) + "'"
}

func LoadDBConfig() (*DBConfig, error) {
	dbHost, err := requireEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	dbUser, err := requireEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	dbPassword, err := requireEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dbName, err := requireEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	maxOpenConnections, err := getIntEnv("DB_MAX_OPEN_CONNECTIONS", 25)
	if err != nil {
		return nil, err
	}
	maxIdleConnections, err := getIntEnv("DB_MAX_IDLE_CONNECTIONS", 10)
	if err != nil {
		return nil, err
	}
	maxLifetimeMinutes, err := getIntEnv("DB_MAX_LIFETIME_MINUTES", 30)
	if err != nil {
		return nil, err
	}

	sslRootCert := getEnv("DB_CA_CERT", "")
	if sslRootCert != "" {
		if err := validateCACert(sslRootCert); err != nil {
			return nil, fmt.Errorf("DB_CA_CERT: %w", err)
		}
	}

	return &DBConfig{
		Host:               dbHost,
		Port:               getEnv("DB_PORT", "5432"),
		User:               dbUser,
		Password:           dbPassword,
		Name:               dbName,
		SSLMode:            getEnv("DB_SSL_MODE", "disable"),
		SSLRootCert:        sslRootCert,
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
		MaxLifetimeMinutes: maxLifetimeMinutes,
	}, nil
}
