package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezect/url-shortener/internal/models"
)

type ClickRepository struct {
	url  string
	pool *pgxpool.Pool
	conn Querier
}

func NewClickRepository(ctx context.Context, url string, pool *pgxpool.Pool) *ClickRepository {
	return &ClickRepository{
		url:  url,
		pool: pool,
		conn: pool,
	}
}

func (db *ClickRepository) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *ClickRepository) WithTx(tx pgx.Tx) *ClickRepository {
	dbCopy := *db
	dbCopy.conn = tx
	return &dbCopy
}

func (r *ClickRepository) Create(ctx context.Context, shortCode string, ip string, userAgent, referer *string) error {
	var ua, ref sql.NullString

	if userAgent != nil {
		ua.String = *userAgent
		ua.Valid = true
	}

	if referer != nil {
		ref.String = *referer
		ref.Valid = true
	}

	_, err := r.conn.Exec(
		ctx,
		"INSERT INTO clicks (short_code, ip, user_agent, referer) VALUES	($1, $2, $3, $4)",
		shortCode,
		ip,
		ua,
		ref,
	)
	if err != nil {
		return fmt.Errorf("Error while inserting values: %v", err.Error())
	}

	return nil
}

func (r *ClickRepository) GetTotalClicks(ctx context.Context, shortCode string) (int64, error) {
	var totalClicks int64

	err := r.conn.QueryRow(ctx, "SELECT COUNT(*) FROM clicks WHERE short_code = $1", shortCode).Scan(&totalClicks)
	if err != nil {
		return 0, err
	}

	return totalClicks, nil
}

func (r *ClickRepository) GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error) {
	rows, err := r.conn.Query(
		ctx,
		`SELECT COUNT(*) as total_clicks, DATE(clicked_at) as date FROM clicks WHERE short_code = $1 GROUP BY DATE(clicked_at) ORDER BY DATE(created_at) DESC`,
		shortCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := make(map[time.Time]int)
	for rows.Next() {
		var totalDayClicks int
		var date time.Time
		if err := rows.Scan(&totalDayClicks, &date); err != nil {
			return nil, err
		}

		report[date] = totalDayClicks
	}

	return &report, nil
}

func (r *ClickRepository) BatchInsert(ctx context.Context, clicks []models.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, click := range clicks {
		query := `INSERT INTO clicks (short_code, ip, user_agent, referer) VALUES ($1, $2, $3, $4) RETURNING id`
		batch.Queue(query, click.ShortCode, click.Ip, click.UserAgent, click.Referer)
	}

	br := r.conn.SendBatch(ctx, batch)
	defer br.Close()

	var ids []int64
	for i := range clicks {
		var id int64
		err := br.QueryRow().Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to scan row %d: %w", i, err)
		}
		ids = append(ids, id)
		clicks[i].Id = id
	}

	return nil
}

func (r *ClickRepository) Stop() {
	r.pool.Close()
}
