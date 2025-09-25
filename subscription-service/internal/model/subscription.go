// internal/model/subscription.go
package model

import "time"

type Subscription struct {
	ID          string     `json:"id"`                      // uuid as string
	ServiceName string     `json:"service_name"`            // название сервиса
	Price       int        `json:"price"`                   // целые рубли
	UserID      string     `json:"user_id"`                 // uuid пользователя
	StartDate   time.Time  `json:"start_date"`              // первый день месяца (YYYY-MM-01)
	EndDate     *time.Time `json:"end_date,omitempty"`      // optional
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
}
