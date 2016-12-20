package db

import (
	"database/sql"
	"fmt"

	"github.com/Clever/analytics-pipeline-monitor/config"
	l "github.com/Clever/analytics-pipeline-monitor/logger"
	_ "github.com/lib/pq" // Postgres driver.
)

// RedshiftClient contains the redshift client connection.
type RedshiftClient struct {
	session *sql.DB
}

// RedshiftCredentials contains the redshift credentials/informatio.
type RedshiftCredentials struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// NewRedshiftClient creates a Redshift db client.
func newRedshiftClient(info RedshiftCredentials) (*RedshiftClient, error) {
	const connectionTimeout = 60
	connectionParams := fmt.Sprintf("host=%s port=%s dbname=%s connect_timeout=%d",
		info.Host, info.Port, info.Database, connectionTimeout)
	credentialsParams := fmt.Sprintf("user=%s password=%s", info.Username, info.Password)

	l.GetKVLogger().InfoD("New-redshift-client", l.M{
		"connectionParams": connectionParams,
	})
	openParams := fmt.Sprintf("%s %s", connectionParams, credentialsParams)
	session, err := sql.Open("postgres", openParams)
	if err != nil {
		return nil, err
	}

	return &RedshiftClient{session}, nil
}

// NewRedshiftProdClient initializes a client to fresh prod
func NewRedshiftProdClient() (*RedshiftClient, error) {
	info := RedshiftCredentials{
		Host:     config.RedshiftProdHost,
		Port:     config.RedshiftProdPort,
		Username: config.RedshiftProdUsername,
		Password: config.RedshiftProdPassword,
		Database: config.RedshiftProdDatabase,
	}

	return newRedshiftClient(info)
}

// NewRedshiftFastClient initializes a client to fast prod
func NewRedshiftFastClient() (*RedshiftClient, error) {
	info := RedshiftCredentials{
		Host:     config.RedshiftFastHost,
		Port:     config.RedshiftFastPort,
		Username: config.RedshiftFastUsername,
		Password: config.RedshiftFastPassword,
		Database: config.RedshiftFastDatabase,
	}

	return newRedshiftClient(info)
}

// QueryLatency returns the latency for a given table,
// defined as the time difference in hours between now
// and the most recent record in a table. Returns the latency,
// if applicable, and whether or not the table contains rows
func (c *RedshiftClient) QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error) {
	latencyQuery := fmt.Sprintf("SELECT datediff(hour, max(%s), getdate()) FROM %s.%s",
		timestampColumn, schemaName, tableName)
	rows, err := c.session.Query(latencyQuery)
	if err != nil {
		return 0, false, err
	}

	var latency sql.NullInt64
	rows.Next()
	if err := rows.Scan(&latency); err != nil {
		return 0, false, err
	}
	return latency.Int64, latency.Valid, nil
}
