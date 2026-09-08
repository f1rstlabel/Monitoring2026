package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"sanoc/backend/internal/domain"
)

type PostgresPublicMonitorGroupRepository struct {
	db *sql.DB
}

func NewPostgresPublicMonitorGroupRepository(db *sql.DB) *PostgresPublicMonitorGroupRepository {
	return &PostgresPublicMonitorGroupRepository{db: db}
}

func scanPublicMonitorGroup(row rowScanner) (*domain.PublicMonitorGroup, error) {
	var group domain.PublicMonitorGroup
	var createdAt, updatedAt time.Time
	if err := row.Scan(&group.ID, &group.Name, &group.Description, &group.DisplayOrder, &group.Enabled, &group.MonitorCount, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	group.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	group.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return &group, nil
}

const publicMonitorGroupColumns = `g.id, g.name, COALESCE(g.description, ''), g.display_order, g.enabled,
    (SELECT COUNT(*) FROM public_monitors m WHERE m.group_id = g.id AND m.deleted_at IS NULL AND m.purged_at IS NULL), g.created_at, g.updated_at`

func (r *PostgresPublicMonitorGroupRepository) GetAll() ([]domain.PublicMonitorGroup, error) {
	if r == nil || r.db == nil {
		return []domain.PublicMonitorGroup{}, nil
	}
	rows, err := r.db.Query("SELECT " + publicMonitorGroupColumns + " FROM public_monitor_groups g ORDER BY g.display_order ASC, g.name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]domain.PublicMonitorGroup, 0)
	for rows.Next() {
		group, err := scanPublicMonitorGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *group)
	}
	return groups, rows.Err()
}

func (r *PostgresPublicMonitorGroupRepository) GetByID(id string) (*domain.PublicMonitorGroup, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrNoRows
	}
	return scanPublicMonitorGroup(r.db.QueryRow("SELECT "+publicMonitorGroupColumns+" FROM public_monitor_groups g WHERE g.id=$1", id))
}

func (r *PostgresPublicMonitorGroupRepository) Create(group *domain.PublicMonitorGroup) (*domain.PublicMonitorGroup, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("public monitor group database is not initialized")
	}
	if group.ID == "" {
		group.ID = publicMonitorGroupID()
	}
	_, err := r.db.Exec(`INSERT INTO public_monitor_groups (id, name, description, display_order, enabled)
		VALUES ($1,$2,$3,$4,$5)`, group.ID, strings.TrimSpace(group.Name), strings.TrimSpace(group.Description), group.DisplayOrder, group.Enabled)
	if err != nil {
		return nil, err
	}
	return r.GetByID(group.ID)
}

func (r *PostgresPublicMonitorGroupRepository) Update(group *domain.PublicMonitorGroup) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor group database is not initialized")
	}
	_, err := r.db.Exec(`UPDATE public_monitor_groups SET name=$1, description=$2, display_order=$3, enabled=$4,
		updated_at=CURRENT_TIMESTAMP WHERE id=$5`, strings.TrimSpace(group.Name), strings.TrimSpace(group.Description), group.DisplayOrder, group.Enabled, group.ID)
	return err
}

func (r *PostgresPublicMonitorGroupRepository) Delete(id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("public monitor group database is not initialized")
	}
	_, err := r.db.Exec("DELETE FROM public_monitor_groups WHERE id=$1 AND name <> 'Ungrouped'", id)
	return err
}

func publicMonitorGroupID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("grp-%d", time.Now().UnixNano())
	}
	return "grp-" + hex.EncodeToString(buf)
}
