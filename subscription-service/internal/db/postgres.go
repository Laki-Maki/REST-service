// internal/db/postgres.go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"subscription-service/internal/model"
)

type Store struct {
	DB *sql.DB
}

// CreateSubscription inserts subscription and returns generated id.
func (s *Store) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
	          VALUES ($1, $2, $3, $4, $5)
			  RETURNING id, created_at, updated_at`
	return s.DB.QueryRowContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (s *Store) GetSubscriptionByID(ctx context.Context, id string) (*model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	          FROM subscriptions WHERE id = $1`
	row := s.DB.QueryRowContext(ctx, query, id)
	var sub model.Subscription
	var endDate sql.NullTime
	if err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
		return nil, err
	}
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	return &sub, nil
}

func (s *Store) DeleteSubscription(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

func (s *Store) UpdateSubscription(ctx context.Context, sub *model.Subscription) error {
	query := `UPDATE subscriptions
	          SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5, updated_at = now()
			  WHERE id = $6
			  RETURNING updated_at`
	return s.DB.QueryRowContext(ctx, query,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID,
	).Scan(&sub.UpdatedAt)
}

// ListSubscriptions: simple list with optional filters (user_id, service_name)
func (s *Store) ListSubscriptions(ctx context.Context, userID *string, serviceName *string, limit, offset int) ([]model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
			  FROM subscriptions
			  WHERE ($1::uuid IS NULL OR user_id = $1)
			    AND ($2::text IS NULL OR service_name = $2)
			  ORDER BY created_at DESC
			  LIMIT $3 OFFSET $4`
	rows, err := s.DB.QueryContext(ctx, query, userID, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, err
		}
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

// FindSubscriptionsOverlapping находит подписки, пересекающиеся с периодом [from, to].
// userID и serviceName могут быть nil (не фильтровать).
func (s *Store) FindSubscriptionsOverlapping(ctx context.Context, from, to time.Time, userID *string, serviceName *string) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
	`
	args := []interface{}{from, to}
	i := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, *userID)
		i++
	}

	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", i)
		args = append(args, *serviceName)
		i++
	}

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("FindSubscriptionsOverlapping query failed: %v\nSQL: %s\nArgs: %+v", err, query, args)
		return nil, err
	}
	defer rows.Close()

	var out []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime
		if err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		out = append(out, sub)
	}

	return out, rows.Err()
}
