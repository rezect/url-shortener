package testhelpers

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rezect/url-shortener/migrations"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type PostgresContainer struct {
	Ctr        *postgres.PostgresContainer
	ConnString string
}

func CreatePostgresContainer(t *testing.T, ctx context.Context) *PostgresContainer {
	// 1. Запускаем PostgreSQL контейнер
	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	// Регистрируем автоматическую очистку контейнера
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// 2. Получаем строку подключения
	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	applyMigrations(t, connStr)

	return &PostgresContainer{
		Ctr:        ctr,
		ConnString: connStr,
	}
}

func applyMigrations(t *testing.T, connStr string) {
	// Открываем соединение с БД
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Устанавливаем диалект (postgres)
	require.NoError(t, goose.SetDialect("postgres"))

	// Применяем миграции
	goose.SetBaseFS(migrations.FS)
	err = goose.Up(db, ".")
	require.NoError(t, err)
}
