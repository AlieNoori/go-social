package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `
	SELECT id,name,level,Description
	FROM roles
	WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(ctx, queryTimeoutDuration)
	defer cancel()

	role := &Role{Name: roleName}

	err := s.db.QueryRowContext(ctx, query, roleName).Scan(&role.ID, &role.Name, &role.Level, &role.Description)
	if err != nil {
		return nil, err
	}

	return role, nil
}
