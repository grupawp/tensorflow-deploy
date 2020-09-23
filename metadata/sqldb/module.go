package sqldb

import (
	"context"
	"database/sql"
	"time"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/metadata"
)

// Module ...
type Module struct {
	connection *sql.DB
}

// Get gets single module metadata
func (m *Module) Get(ctx context.Context, parameters app.QueryParameters) (*app.ModuleData, error) {
	module := new(app.ModuleData)

	query, queryValues, err := buildExtendedSearchQueryAndValues("SELECT id, team, project, name, version, created, updated FROM module", "LIMIT 1", parameters)
	if err != nil {
		return nil, err
	}

	err = m.connection.QueryRowContext(ctx, query, queryValues...).Scan(&module.ID, &module.Team, &module.Project, &module.Name, &module.Version, &module.Created, &module.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, exterr.WrapWithFrame(err)
	}

	return module, nil
}

// Add inserts module metadata
func (m *Module) Add(ctx context.Context, module app.ModuleData) (int64, error) {
	timestamp := time.Now().Unix()
	result, err := m.connection.ExecContext(ctx, "INSERT INTO module (team, project, name, version, created, updated) VALUES(?, ?, ?, ?, ?, ?)",
		module.Team,
		module.Project,
		module.Name,
		module.Version,
		timestamp,
		timestamp)
	if err != nil {
		return metadata.InvalidID, exterr.WrapWithFrame(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return metadata.InvalidID, exterr.WrapWithFrame(err)
	}

	return id, nil
}

// Delete deletes module metadata
func (m *Module) Delete(ctx context.Context, id int64) error {
	result, err := m.connection.ExecContext(ctx, "DELETE FROM module WHERE id = ?", id)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return exterr.WrapWithFrame(err)
	} else if affected == 0 {
		return errorDeleteModule
	}

	return nil
}

// NextVersion returns next available module version
func (m *Module) NextVersion(ctx context.Context, parameters app.QueryParameters) (int64, error) {
	query, queryValues, err := buildExtendedSearchQueryAndValues("SELECT version FROM module", "ORDER BY version DESC LIMIT 1", parameters)
	if err != nil {
		return metadata.InvalidVersion, err
	}

	var version int64

	err = m.connection.QueryRowContext(ctx, query, queryValues...).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return metadata.StartVersion, nil
		}
		return metadata.InvalidVersion, exterr.WrapWithFrame(err)
	}

	return metadata.NextVersion(version)
}

// List modules metadata
func (m *Module) List(ctx context.Context, parameters app.QueryParameters) ([]*app.ModuleData, error) {
	query, queryValues, err := buildSearchQueryAndValues("SELECT id, team, project, name, version, created, updated FROM module", parameters)
	if err != nil {
		return nil, err
	}

	modules := make([]*app.ModuleData, 0)

	rows, err := m.connection.QueryContext(ctx, query, queryValues...)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	defer rows.Close()

	for rows.Next() {
		module := new(app.ModuleData)

		if err := rows.Scan(&module.ID, &module.Team, &module.Project, &module.Name, &module.Version, &module.Created, &module.Updated); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}

		modules = append(modules, module)
	}
	if rows.Err() != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if len(modules) == 0 {
		return nil, nil
	}

	return modules, nil
}

// ListUniqueTeamProject lists distinct keys (team, project)
func (m *Module) ListUniqueTeamProject(ctx context.Context) ([]*app.ServableID, error) {
	servables := make([]*app.ServableID, 0)

	rows, err := m.connection.QueryContext(ctx, "SELECT team, project FROM module GROUP BY team, project ORDER BY team, project")
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	defer rows.Close()

	for rows.Next() {
		servable := new(app.ServableID)

		if err := rows.Scan(&servable.Team, &servable.Project); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}

		servables = append(servables, servable)
	}
	if rows.Err() != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if len(servables) == 0 {
		return nil, nil
	}

	return servables, nil
}
