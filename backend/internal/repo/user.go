package repo

import (
	"context"
	"fmt"

	"yardpass/internal/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*PostgresRepo
}

func NewUserRepo(repo *PostgresRepo) *UserRepo {
	return &UserRepo{repo}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.BuildingID,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.BuildingID,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, building_id, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.BuildingID,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, role = $5, building_id = $6, status = $7
		WHERE id = $1
		RETURNING updated_at
	`

	return r.pool.QueryRow(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.BuildingID,
		user.Status,
	).Scan(&user.UpdatedAt)
}

func (r *UserRepo) List(ctx context.Context, filters domain.UserFilters) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filters.Role != nil {
		query += fmt.Sprintf(` AND role = $%d`, argPos)
		args = append(args, *filters.Role)
		argPos++
	}

	if filters.BuildingID != nil {
		query += fmt.Sprintf(` AND building_id = $%d`, argPos)
		args = append(args, *filters.BuildingID)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argPos)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.BuildingID,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

