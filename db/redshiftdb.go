package db

import (
	"database/sql"
	"fmt"

	"github.com/Clever/analytics-pipeline-monitor/config"
	"github.com/Clever/analytics-pipeline-monitor/logger"
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

	logger.Info("New-redshift-client", logger.M{
		"connectionParams": connectionParams,
	})
	openParams := fmt.Sprintf("%s %s", connectionParams, credentialsParams)
	session, err := sql.Open("postgres", openParams)
	if err != nil {
		return nil, err
	}

	return &RedshiftClient{session}, nil
}

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

// For testing
func (c *RedshiftClient) CountRows(tablename string) (int64, error) {
	query := fmt.Sprintf("Select count(*) from %s", tablename)
	rows, err := c.session.Query(query)
	if err != nil {
		return 0, err
	}

	var count sql.NullInt64
	rows.Next()
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}

	return count.Int64, nil
}
