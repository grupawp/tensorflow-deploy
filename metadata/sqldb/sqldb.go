package sqldb

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/metadata"
)

var (
	typeNotSupportedErrorCode = 1003
	deleteModelErrorCode      = 1004
	updateModelErrorCode      = 1005
	deleteModuleErrorCode     = 1006

	// General model error messages
	errorUpdateModel      = exterr.NewErrorWithMessage("update model error").WithComponent(app.ComponentMetadata).WithCode(updateModelErrorCode)
	errorDeleteModel      = exterr.NewErrorWithMessage("delete model error").WithComponent(app.ComponentMetadata).WithCode(deleteModelErrorCode)
	errorTypeNotSupported = exterr.NewErrorWithMessage("type not supported").WithComponent(app.ComponentMetadata).WithCode(typeNotSupportedErrorCode)

	// General module error message
	errorDeleteModule = exterr.NewErrorWithMessage("delete module error").WithComponent(app.ComponentMetadata).WithCode(deleteModuleErrorCode)
)

// SQLDB ...
type SQLDB struct {
	Model  *Model
	Module *Module

	driver     string
	connection *sql.DB
}

// NewSQLDB ...
func NewSQLDB(ctx context.Context, driver, dsn string) (*SQLDB, error) {
	connection, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	db := &SQLDB{
		Model:  &Model{connection: connection},
		Module: &Module{connection: connection},

		driver:     driver,
		connection: connection,
	}

	return db, nil
}

// Close closes the connection
func (s *SQLDB) Close(ctx context.Context) (err error) {
	err = s.connection.Close()
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

// buildSearchQueryAndValues builds search query with dynamic "WHERE" part and its values
func buildSearchQueryAndValues(queryPart string, params app.QueryParameters) (query string, values []interface{}, err error) {
	var (
		fields []string
	)

	if len(params) == 0 {
		return queryPart, values, nil
	}

	for field, v := range params {
		switch value := v.(type) {
		case uint8, int64:
			fields = append(fields, field+"=?")
			values = append(values, value)
		case string:
			if field == app.RequestFieldStatus {
				status, err := metadata.StatusToID(value)
				if err != nil {
					return "", nil, exterr.WrapWithFrame(err)
				}
				values = append(values, status)
			} else {
				values = append(values, value)
			}

			fields = append(fields, field+"=?")
		default:
			return "", values, errorTypeNotSupported
		}
	}

	if len(fields) > 0 {
		query = queryPart + " WHERE " + strings.Join(fields, " AND ")
	}

	return query, values, nil
}

// buildExtendedSearchQueryAndValues builds search query with dynamic "WHERE" part and its values
func buildExtendedSearchQueryAndValues(queryPart, queryExtendPart string, params app.QueryParameters) (query string, values []interface{}, err error) {
	query, values, err = buildSearchQueryAndValues(queryPart, params)
	if err != nil {
		return query, values, exterr.WrapWithFrame(err)
	}

	return query + " " + queryExtendPart, values, nil
}
