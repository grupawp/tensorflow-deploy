package sqldb

import (
	"context"
	"database/sql"
	"time"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/grupawp/tensorflow-deploy/metadata"
)

// Model ...
type Model struct {
	connection *sql.DB
}

// Get gets single model metadata
func (m *Model) Get(ctx context.Context, parameters app.QueryParameters) (*app.ModelData, error) {
	model := new(app.ModelData)
	var statusID uint8

	query, queryValues, err := buildExtendedSearchQueryAndValues("SELECT id, team, project, name, version, label, status, created, updated FROM model", "LIMIT 1", parameters)
	if err != nil {
		return nil, err
	}

	err = m.connection.QueryRowContext(ctx, query, queryValues...).Scan(&model.ID, &model.Team, &model.Project, &model.Name, &model.Version, &model.Label, &statusID, &model.Created, &model.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, exterr.WrapWithFrame(err)
	}

	model.Status, err = metadata.StatusToName(statusID)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// Add inserts model metadata
func (m *Model) Add(ctx context.Context, model app.ModelData) (int64, error) {
	status, err := metadata.StatusToID(model.Status)
	if err != nil {
		return metadata.InvalidID, err
	}

	timestamp := time.Now().Unix()
	result, err := m.connection.ExecContext(ctx, "INSERT INTO model (team, project, name, version, label, status, created, updated) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		model.Team,
		model.Project,
		model.Name,
		model.Version,
		model.Label,
		status,
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

// UpdateStatus updates model status
func (m *Model) UpdateStatus(ctx context.Context, id int64, status string) error {
	statusID, err := metadata.StatusToID(status)
	if err != nil {
		return err
	}

	result, err := m.connection.ExecContext(ctx, "UPDATE model SET status = ?, updated = ? WHERE id = ? AND status != ?",
		statusID,
		time.Now().Unix(),
		id,
		statusID)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return exterr.WrapWithFrame(err)
	} else if affected == 0 {
		return errorUpdateModel
	}

	return nil
}

// Delete deletes model metadata
func (m *Model) Delete(ctx context.Context, id int64) error {
	result, err := m.connection.ExecContext(ctx, "DELETE FROM model WHERE id = ?", id)
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return exterr.WrapWithFrame(err)
	} else if affected == 0 {
		return errorDeleteModel
	}

	return nil
}

// NextVersion returns next available model version
func (m *Model) NextVersion(ctx context.Context, parameters app.QueryParameters) (int64, error) {
	query, queryValues, err := buildExtendedSearchQueryAndValues("SELECT version FROM model", "ORDER BY version DESC LIMIT 1", parameters)
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

// List lists models metadata
func (m *Model) List(ctx context.Context, parameters app.QueryParameters) ([]*app.ModelData, error) {
	query, queryValues, err := buildSearchQueryAndValues("SELECT id, team, project, name, version, label, status, created, updated FROM model", parameters)
	if err != nil {
		return nil, err
	}

	models := make([]*app.ModelData, 0)

	rows, err := m.connection.QueryContext(ctx, query, queryValues...)
	if err != nil {
		return nil, exterr.WrapWithFrame(err)
	}
	defer rows.Close()

	for rows.Next() {
		model := new(app.ModelData)
		var statusID uint8

		if err := rows.Scan(&model.ID, &model.Team, &model.Project, &model.Name, &model.Version, &model.Label, &statusID, &model.Created, &model.Updated); err != nil {
			return nil, exterr.WrapWithFrame(err)
		}

		model.Status, err = metadata.StatusToName(statusID)
		if err != nil {
			return nil, err
		}

		models = append(models, model)
	}
	if rows.Err() != nil {
		return nil, exterr.WrapWithFrame(err)
	}

	if len(models) == 0 {
		return nil, nil
	}

	return models, nil
}

// ListUniqueTeamProject lists distinct keys (team, project)
func (m *Model) ListUniqueTeamProject(ctx context.Context) ([]*app.ServableID, error) {
	servables := make([]*app.ServableID, 0)

	rows, err := m.connection.QueryContext(ctx, "SELECT team, project FROM model GROUP BY team, project ORDER BY team, project")
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

// RemoveLabel removes model label from metadata for given ModelData
func (m *Model) RemoveLabel(ctx context.Context, model app.ModelData) error {
	params := app.QueryParameters{"team": model.Team, "project": model.Project,
		"name": model.Name, "label": model.Label}

	currentLabeledModel, err := m.Get(ctx, params)
	if err != nil {
		return err
	}
	return m.Delete(ctx, currentLabeledModel.ID)
}

func (m *Model) ChangeLabel(ctx context.Context, model app.ModelData) error {
	params := app.QueryParameters{"team": model.Team, "project": model.Project,
		"name": model.Name, "label": model.Label}

	currentLabeledModel, err := m.Get(ctx, params)
	if err != nil {
		return err
	}

	tx, err := m.connection.Begin()
	if err != nil {
		return exterr.WrapWithFrame(err)
	}

	if currentLabeledModel != nil {
		result, err := tx.ExecContext(ctx, "DELETE FROM model WHERE id = ?", currentLabeledModel.ID)
		if err != nil {
			tx.Rollback()

			return exterr.WrapWithFrame(err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()

			return exterr.WrapWithFrame(err)
		} else if affected == 0 {
			tx.Rollback()

			return errorUpdateModel
		}
	}

	status, err := metadata.StatusToID(model.Status)
	if err != nil {
		tx.Rollback()

		return err
	}

	timestamp := time.Now().Unix()
	_, err = tx.ExecContext(ctx, "INSERT INTO model (team, project, name, version, label, status, created, updated) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		model.Team,
		model.Project,
		model.Name,
		model.Version,
		model.Label,
		status,
		timestamp,
		timestamp)
	if err != nil {
		tx.Rollback()

		return exterr.WrapWithFrame(err)
	}

	if err := tx.Commit(); err != nil {
		return exterr.WrapWithFrame(err)
	}

	return nil
}

// IsStatusPending checks if status is set to pending
func (m *Model) IsStatusPending(ctx context.Context, servableID app.ServableID) (bool, error) {
	return m.isStatusSet(ctx, servableID, metadata.StatusPending)
}

// isStatusSet checks if given status is set
func (m *Model) isStatusSet(ctx context.Context, servableID app.ServableID, status uint8) (bool, error) {
	var statusID uint8

	err := m.connection.QueryRowContext(ctx, "SELECT status FROM model WHERE team = ? AND project = ? AND name = ? AND status = ? LIMIT 1", servableID.Team, servableID.Project, servableID.Name, status).Scan(&statusID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, exterr.WrapWithFrame(err)
	}

	return true, nil
}
