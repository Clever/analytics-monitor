package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Clever/analytics-monitor/config"
	l "github.com/Clever/analytics-monitor/logger"
	// Use our own version of the postgres library so we get keep-alive support.
	// See https://github.com/Clever/pq/pull/1
	_ "github.com/Clever/pq"
)

// PostgresClient exposes an interface for querying Postgres.
type PostgresClient interface {
	GetClusterName() string
	QueryTableMetadata(schemaName string) (map[string]TableMetadata, error)
	QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error)
	QuerySTLLoadErrors() ([]LoadError, error)
}

// postgresClient provides a default implementation of PostgresClient
// that contains the postgres client connection.
type postgresClient struct {
	session     *sql.DB
	clusterName string
}

// PostgresCredentials contains the postgres credentials/information.
type PostgresCredentials struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// TableMetadata contains information about a table in Postgres
type TableMetadata struct {
	TableName       string
	TimestampColumn string
}

// LoadError contains information surfacing load errors
type LoadError struct {
	TableNames string `json:"table_names"`
	ErrorCode  int64  `json:"error_code"`
	Count      int64  `json:"count"`
}

// NewPostgresClient creates a Postgres db client.
func newPostgresClient(info PostgresCredentials, clusterName string) (PostgresClient, error) {
	const connectionTimeout = 60
	connectionParams := fmt.Sprintf("host=%s port=%s dbname=%s keepalive=1 connect_timeout=%d",
		info.Host, info.Port, info.Database, connectionTimeout)
	credentialsParams := ""
	if len(info.Username) > 0 {
		credentialsParams = fmt.Sprintf("user=%s password=%s", info.Username, info.Password)
	} else {
		// Locally we have to disable ssl mode
		credentialsParams = "sslmode=disable"
	}

	l.GetKVLogger().InfoD("New-postgres-client", l.M{
		"connectionParams": connectionParams,
	})
	openParams := fmt.Sprintf("%s %s", connectionParams, credentialsParams)
	session, err := sql.Open("postgres", openParams)
	if err != nil {
		return nil, err
	}

	return &postgresClient{session, clusterName}, nil
}

// NewPostgresClient initializes a postgres client
func NewPostgresClient() (PostgresClient, error) {
	info := PostgresCredentials{
		Host:     config.PostgresHost,
		Port:     config.PostgresPort,
		Username: config.PostgresUsername,
		Password: config.PostgresPassword,
		Database: config.PostgresDatabase,
	}

	return newPostgresClient(info, "redshift-prod")
}

// GetClusterName returns the name of the client Postgres cluster
func (c *postgresClient) GetClusterName() string {
	return c.clusterName
}

// QueryTableMetadata returns a map of tables
// belonging to a given schema in Postgres, indexed
// by table name.
// It also attempts to infer the timestamp column, by
// choosing the alphabetically lowest column with a
// timestamp type. We use this as a heuristic since a
// lot of our timestamp columns are prefixed with "_".
func (c *postgresClient) QueryTableMetadata(schemaName string) (map[string]TableMetadata, error) {
	query := fmt.Sprintf(`
		SELECT table_name, min("column_name")
		FROM information_schema.columns
		WHERE table_schema = '%s'
		AND data_type ILIKE '%%timestamp%%'
		GROUP BY table_name
	`, schemaName)

	tableMetadata := make(map[string]TableMetadata)
	rows, err := c.session.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row TableMetadata
		if err := rows.Scan(&row.TableName, &row.TimestampColumn); err != nil {
			return tableMetadata, fmt.Errorf("Unable to scan row for schema %s: %s", schemaName, err)
		}

		tableMetadata[row.TableName] = row
	}

	return tableMetadata, nil
}

// QueryLatency returns the latency for a given table,
// defined as the time difference in hours between now
// and the most recent record in a table. Returns the latency,
// if applicable, and whether or not the table contains rows
func (c *postgresClient) QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error) {
	// We extract the epoch because it works in both Redshift and Postgres
	latencyQuery := fmt.Sprintf("SELECT extract(epoch from MAX(\"%s\")) FROM \"%s\".\"%s\"", timestampColumn, schemaName, tableName)
	rows, err := c.session.Query(latencyQuery)
	if err != nil {
		return 0, false, fmt.Errorf("Error executing query %s: %s", latencyQuery, err)
	}

	var latency sql.NullFloat64
	rows.Next()
	if err := rows.Scan(&latency); err != nil {
		return 0, false, fmt.Errorf("Unable to scan row for query %s: %s", latencyQuery, err)
	}
	hourDiff := (time.Now().Unix() - int64(latency.Float64)) / 3600
	return hourDiff, latency.Valid, nil
}

func (c *postgresClient) QuerySTLLoadErrors() ([]LoadError, error) {
	query := fmt.Sprintf(`
		SELECT sum("count") AS count, err_code, listagg(name, ', ')
    FROM (SELECT COUNT(stl.err_code) AS count, stl.err_code, stv.name
    FROM stl_load_errors AS stl
    INNER JOIN stv_tbl_perm AS stv ON stl.tbl = stv.id
    WHERE starttime > (getdate() - INTERVAL '3 hour')
        AND filename not like 's3://firehose-prod/github-events%%'
    GROUP BY name, err_code)
    GROUP BY err_code
  `)

	var loadErrors []LoadError
	rows, err := c.session.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var row LoadError
		if err := rows.Scan(&row.Count, &row.ErrorCode, &row.TableNames); err != nil {
			return loadErrors, fmt.Errorf("Unable to scan row: %s", err)
		}

		loadErrors = append(loadErrors, row)
	}

	return loadErrors, nil
}
