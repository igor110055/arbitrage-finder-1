package postgres

import (
	"net/url"
	"time"
)

const (
	connMaxLifetime = 5 * time.Minute
)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
	CertPath   string
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit.
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxOpenConns int
	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
	// MaxOpenConns limit.
	//
	// If n <= 0, then there is no limit on the number of open connections.
	// The default is 0 (unlimited).
	MaxIdleConns int
}

// URL returns database config in URL presentation
func (c Config) URL() *url.URL {
	sslMode := "verify-full"
	if c.DisableTLS {
		sslMode = "disable"
	}
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	if !c.DisableTLS {
		q.Set("sslrootcert", c.CertPath)
	}
	q.Set("timezone", "utc")

	return &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     c.Host,
		Path:     c.Name,
		RawQuery: q.Encode(),
	}
}
