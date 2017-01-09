package db

import (
	"database/sql"
	"fmt"

	"github.com/Clever/analytics-pipeline-monitor/config"
	l "github.com/Clever/analytics-pipeline-monitor/logger"
	_ "github.com/lib/pq" // Postgres driver.
)

// RedshiftClient exposes an interface for querying Redshift.
type RedshiftClient interface {
	QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error)
	QuerySTLLoadErrors() ([]LoadError, error)
}

// redshiftClient provides a default implementation of RedshiftClient
// that contains the redshift client connection.
type redshiftClient struct {
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

type LoadError struct {
	TableNames string `json:"table_names"`
	ErrorCode  int64  `json:"error_code"`
	Count      int64  `json:"count"`
}

// NewRedshiftClient creates a Redshift db client.
func newRedshiftClient(info RedshiftCredentials) (RedshiftClient, error) {
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

	return &redshiftClient{session}, nil
}

// NewRedshiftProdClient initializes a client to fresh prod
func NewRedshiftProdClient() (RedshiftClient, error) {
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
func NewRedshiftFastClient() (RedshiftClient, error) {
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
func (c *redshiftClient) QueryLatency(timestampColumn, schemaName, tableName string) (int64, bool, error) {
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

func (c *redshiftClient) QuerySTLLoadErrors() ([]LoadError, error) {
	query := fmt.Sprintf(`
		SELECT sum("count") AS count, err_code, listagg(name, ', ')
    FROM (SELECT COUNT(stl.err_code) AS count, stl.err_code, stv.name 
    FROM stl_load_errors AS stl 
    INNER JOIN stv_tbl_perm AS stv ON stl.tbl = stv.id
    WHERE starttime > (getdate() - INTERVAL '3 hour')
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
		} else {
			loadErrors = append(loadErrors, row)
		}
	}

	return loadErrors, nil
}
