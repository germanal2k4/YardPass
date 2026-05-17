package repo

import (
	"context"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*PostgresRepo
}

func NewUserRepo(repo *PostgresRepo) *UserRepo {
	return &UserRepo{repo}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	ctx = queryNameToContext(ctx, "UserRepo.GetByID")
	row, err := r.queries.GetUserByID(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return userFromGetByIDRow(row), nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx = queryNameToContext(ctx, "UserRepo.GetByUsername")
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return userFromGetByUsernameRow(row), nil
}

func (r *UserRepo) GetByNormalizedEmail(ctx context.Context, normalizedEmail string) (*domain.User, error) {
	ctx = queryNameToContext(ctx, "UserRepo.GetByNormalizedEmail")
	if normalizedEmail == "" {
		return nil, nil
	}
	row, err := r.queries.GetUserByNormalizedEmail(ctx, normalizedEmail)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return userFromNormalizedEmailRow(row), nil
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	ctx = queryNameToContext(ctx, "UserRepo.Create")
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Username:        user.Username,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		Role:            user.Role,
		BuildingID:      user.BuildingID,
		ApartmentNumber: user.ApartmentNumber,
		Status:          user.Status,
	})
	if err != nil {
		return err
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	ctx = queryNameToContext(ctx, "UserRepo.Update")
	updatedAt, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		Role:            user.Role,
		BuildingID:      user.BuildingID,
		ApartmentNumber: user.ApartmentNumber,
		Status:          user.Status,
	})
	if err != nil {
		return err
	}
	user.UpdatedAt = updatedAt
	return nil
}

func (r *UserRepo) List(ctx context.Context, filters domain.UserFilters) ([]domain.User, error) {
	ctx = queryNameToContext(ctx, "UserRepo.List")
	rows, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		FilterRole:       filters.Role,
		FilterBuildingID: filters.BuildingID,
		FilterStatus:     filters.Status,
		MaxResults:       intToInt32Ptr(filters.Limit),
		ResultsOffset:    intToInt32Ptr(filters.Offset),
	})
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, len(rows))
	for i, row := range rows {
		users[i] = domain.User(row)
	}
	return users, nil
}

func userFromNormalizedEmailRow(row db.GetUserByNormalizedEmailRow) *domain.User {
	u := domain.User{
		ID:              row.ID,
		Username:        row.Username,
		Email:           row.Email,
		PasswordHash:    row.PasswordHash,
		Role:            row.Role,
		BuildingID:      row.BuildingID,
		ApartmentNumber: row.ApartmentNumber,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	return &u
}

func userFromGetByIDRow(row db.GetUserByIDRow) *domain.User {
	res := domain.User(row)
	return &res
}

func userFromGetByUsernameRow(row db.GetUserByUsernameRow) *domain.User {
	res := domain.User(row)
	return &res
}
