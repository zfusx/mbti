package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zfusx/mbti/internal/models"
	"github.com/zfusx/mbti/internal/scoring"
)

var (
	// ErrNotFound is returned when the requested row does not exist.
	ErrNotFound = errors.New("record not found")
)

// Storage wraps access to PostgreSQL.
type Storage struct {
	pool *pgxpool.Pool
}

// New creates a pgx connection pool.
func New(ctx context.Context, connString string) (*Storage, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Storage{pool: pool}, nil
}

// Close shuts down the pool.
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Ping verifies that PostgreSQL is reachable.
func (s *Storage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// GetQuestions returns question payloads.
func (s *Storage) GetQuestions(ctx context.Context, limit int, randomize bool) ([]models.Question, error) {
	rankOrder := "id"
	resultOrder := "id"
	if randomize {
		rankOrder = "random()"
		resultOrder = "random()"
	}

	query := fmt.Sprintf(`
		SELECT payload
		FROM (
			SELECT id, payload,
				row_number() OVER (PARTITION BY dimension ORDER BY %s) AS dimension_rank
			FROM mbti_questions
		) ranked
		WHERE dimension_rank <= $1
		ORDER BY %s`, rankOrder, resultOrder)
	rows, err := s.pool.Query(ctx, query, limit/4)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	var items []models.Question
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}

		var q models.Question
		if err := json.Unmarshal(data, &q); err != nil {
			return nil, fmt.Errorf("decode question json: %w", err)
		}
		items = append(items, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questions: %w", err)
	}

	return items, nil
}

// GetQuestionsByIDs returns the question definitions for a submitted answer set.
func (s *Storage) GetQuestionsByIDs(ctx context.Context, ids []int32) ([]models.Question, error) {
	rows, err := s.pool.Query(ctx, "SELECT payload FROM mbti_questions WHERE id = ANY($1::int[])", ids)
	if err != nil {
		return nil, fmt.Errorf("query questions by ids: %w", err)
	}
	defer rows.Close()

	items := make([]models.Question, 0, len(ids))
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		var q models.Question
		if err := json.Unmarshal(data, &q); err != nil {
			return nil, fmt.Errorf("decode question json: %w", err)
		}
		items = append(items, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate questions: %w", err)
	}
	return items, nil
}

// GetResult fetches a type narrative from storage.
func (s *Storage) GetResult(ctx context.Context, mbtiType string) (models.ResultDocument, error) {
	row := s.pool.QueryRow(ctx, "SELECT payload FROM mbti_results WHERE type = $1", mbtiType)

	var data []byte
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ResultDocument{}, ErrNotFound
		}
		return models.ResultDocument{}, fmt.Errorf("scan result: %w", err)
	}

	var doc models.ResultDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return models.ResultDocument{}, fmt.Errorf("decode result json: %w", err)
	}

	return doc, nil
}

// CreateSession inserts an empty session and returns its ID.
func (s *Storage) CreateSession(ctx context.Context) (uuid.UUID, error) {
	row := s.pool.QueryRow(ctx, "INSERT INTO mbti_sessions (answers) VALUES ($1::jsonb) RETURNING id", "[]")

	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert session: %w", err)
	}

	return id, nil
}

// SaveSessionAnswers stores the submitted answers and computed result.
func (s *Storage) SaveSessionAnswers(ctx context.Context, sessionID uuid.UUID, answers []models.Answer, result scoring.Result) error {
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return fmt.Errorf("marshal answers: %w", err)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	tag, err := s.pool.Exec(
		ctx,
		`UPDATE mbti_sessions SET answers = $1::jsonb, result = $2::jsonb WHERE id = $3`,
		string(answersJSON),
		string(resultJSON),
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
