package main

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/config"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/logging"
)

const (
	logConfigErrorCode            = "1001"
	logConfParseErrorCode         = "1002"
	logMetadataBootstrapErrorCode = "1003"
	logcreateTablesErrorCode      = "1004"
)

func main() {
	ctx := context.Background()
	conf, err := config.NewConfig(ctx)
	if err != nil {
		if errors.Is(err, app.ErrCLIUsage) {
			return
		}
		logging.FatalErrorWithStack(ctx, exterr.WrapWithFrame(err), logConfigErrorCode)

	}
	mainConfig, err := conf.Parse(ctx)
	if err != nil {
		logging.FatalErrorWithStack(ctx, exterr.WrapWithFrame(err), logConfParseErrorCode)
	}

	md, err := newMetadataBootstrap(ctx, *mainConfig.Metadata.SQLDB.Driver, *mainConfig.Metadata.SQLDB.DSN)
	if err != nil {
		logging.FatalErrorWithStack(ctx, exterr.WrapWithFrame(err), logMetadataBootstrapErrorCode)
	}

	if err := md.createTables(ctx, tableSQLiteModelDefinition, tableSQLiteModuleDefinition); err != nil {
		logging.FatalErrorWithStack(ctx, exterr.WrapWithFrame(err), logcreateTablesErrorCode)
	}
}

const (
	tableSQLiteModelDefinition = `CREATE TABLE IF NOT EXISTS model (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team VARCHAR(250) NOT NULL,
		project VARCHAR(250) NOT NULL,
		name VARCHAR(250) NOT NULL,
		version INTEGER NOT NULL,
		label VARCHAR(250) NOT NULL,
		status INTEGER NOT NULL,
		created INTEGER NOT NULL,
		updated INTEGER NOT NULL);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_model ON model (team, project, name, version, label);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_model_label ON model (team, project, name, label) WHERE label!="";`

	tableSQLiteModuleDefinition = `CREATE TABLE IF NOT EXISTS module (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		team VARCHAR(250) NOT NULL,
		project VARCHAR(250) NOT NULL,
		name VARCHAR(250) NOT NULL,
		version INTEGER NOT NULL,
		created INTEGER NOT NULL,
		updated INTEGER NOT NULL);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_module ON module (team, project, name, version);`
)

type metadataBootstrap struct {
	connection *sql.DB
}

func newMetadataBootstrap(ctx context.Context, driver, dsn string) (*metadataBootstrap, error) {
	connection, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	return &metadataBootstrap{
		connection: connection,
	}, nil
}

// createTables creates tables from the given schemas
func (b *metadataBootstrap) createTables(ctx context.Context, schemas ...string) error {
	if len(schemas) == 0 {
		return errors.New("no schemas provided")
	}

	for _, schema := range schemas {
		if err := b.createTable(ctx, schema); err != nil {
			return err
		}
	}

	return nil
}

// createTable creates table from the given schema
func (b *metadataBootstrap) createTable(ctx context.Context, schema string) error {
	if _, err := b.connection.ExecContext(ctx, schema); err != nil {
		logging.Error(ctx, err.Error(), "")

		return errors.New("create table error")
	}

	return nil
}
