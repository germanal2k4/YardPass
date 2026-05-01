//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"
	"yardpass/internal/repo"
	"yardpass/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

func TestIntegration_PassFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("yardpass_test"),
		postgres.WithUsername("yardpass"),
		postgres.WithPassword("password"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = postgresContainer.Terminate(ctx)
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Wait for postgres to be ready
	time.Sleep(3 * time.Second)
	runMigrations(t, connStr)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	pgRepo := repo.NewPostgresRepoFromPool(pool, zap.NewNop())
	passRepo := repo.NewPassRepo(pgRepo)
	apartmentRepo := repo.NewApartmentRepo(pgRepo)
	residentRepo := repo.NewResidentRepo(pgRepo)
	ruleRepo := repo.NewRuleRepo(pgRepo)
	scanEventRepo := repo.NewScanEventRepo(pgRepo)

	buildingID := insertTestBuilding(t, ctx, pool)
	apartmentID := insertTestApartment(t, ctx, pool, buildingID)
	residentID := insertTestResident(t, ctx, pool, apartmentID)

	noopMetrics := &metrics.Metrics{}
	passService := service.NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, "test-secret", zap.NewNop(), noopMetrics)

	// Create pass
	now := time.Now().UTC()
	validTo := now.Add(2 * time.Hour)
	req := domain.CreatePassRequest{
		ApartmentID: apartmentID,
		ResidentID:  &residentID,
		CarPlate:    strPtr("A123BC"),
		ValidFrom:   now,
		ValidTo:     validTo,
	}

	pass, err := passService.CreatePass(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, pass)
	assert.NotEmpty(t, pass.ID)
	assert.Equal(t, "A123BC", *pass.CarPlate)
	assert.Equal(t, "active", pass.Status)

	// Validate pass
	result, err := passService.ValidatePass(ctx, pass.ID, 1)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "A123BC", result.CarPlate)

	// Second validation should fail (already used)
	result2, err := passService.ValidatePass(ctx, pass.ID, 1)
	require.NoError(t, err)
	assert.False(t, result2.Valid)
	assert.Equal(t, "PASS_ALREADY_USED", result2.Reason)
}

func strPtr(s string) *string { return &s }

func runMigrations(t *testing.T, connStr string) {
	t.Helper()
	baseDir := findMigrationsDir(t)
	files := []string{
		"001_initial_schema.sql",
		"002_add_building_to_users.sql",
		"003_make_car_plate_optional.sql",
		"004_add_resident_id_to_passes.sql",
	}
	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(context.Background(), connStr)
		if err == nil {
			if err = pool.Ping(context.Background()); err == nil {
				break
			}
			pool.Close()
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)
	defer pool.Close()
	for _, f := range files {
		path := filepath.Join(baseDir, f)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		sql := string(data)
		// Only run Up portion (before -- +goose Down)
		if idx := strings.Index(sql, "\n-- +goose Down"); idx > 0 {
			sql = sql[:idx]
		}
		_, err = pool.Exec(context.Background(), sql)
		require.NoError(t, err, "migration %s", f)
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	for _, dir := range []string{"../../migrations", "../migrations", "migrations"} {
		p := filepath.Join(dir)
		if _, err := os.Stat(filepath.Join(p, "001_initial_schema.sql")); err == nil {
			return p
		}
	}
	t.Fatal("migrations directory not found")
	return ""
}

func insertTestBuilding(t *testing.T, ctx context.Context, pool *pgxpool.Pool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO buildings (name, address) VALUES ('Test', '') RETURNING id`).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertTestApartment(t *testing.T, ctx context.Context, pool *pgxpool.Pool, buildingID int64) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO apartments (building_id, number) VALUES ($1, '101') RETURNING id`, buildingID).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertTestResident(t *testing.T, ctx context.Context, pool *pgxpool.Pool, apartmentID int64) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(ctx, `INSERT INTO residents (apartment_id, telegram_id, chat_id, status) VALUES ($1, 12345, 12345, 'active') RETURNING id`, apartmentID).Scan(&id)
	require.NoError(t, err)
	return id
}
