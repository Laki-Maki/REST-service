// TODO: Реализовать бизнес-логику для подписок
// internal/service/subscription.go
package service

import (
	"time"

	"subscription-service/internal/model"
)

// monthsInclusive returns number of months inclusive between a and b.
// Assumes a and b are normalized to first day of month and a <= b.
func monthsInclusive(a, b time.Time) int {
	years := b.Year() - a.Year()
	months := int(b.Month()) - int(a.Month())
	return years*12 + months + 1
}

// CalculateTotal computes total price (in whole rubles) for provided subscriptions
// over the period [from, to] inclusive.
// It multiplies (number of months of overlap) * price for each subscription.
func CalculateTotal(subs []model.Subscription, from, to time.Time) int64 {
	var total int64 = 0
	for _, s := range subs {
		// left = max(s.StartDate, from)
		left := s.StartDate
		if left.Before(from) {
			left = from
		}
		// right = min(s.EndDateOrInfinite, to)
		right := to
		if s.EndDate != nil && s.EndDate.Before(right) {
			right = *s.EndDate
		}
		// If left > right => no overlap
		if left.After(right) {
			continue
		}
		// normalize to first day of month for calculation
		left = time.Date(left.Year(), left.Month(), 1, 0, 0, 0, 0, time.UTC)
		right = time.Date(right.Year(), right.Month(), 1, 0, 0, 0, 0, time.UTC)
		months := monthsInclusive(left, right)
		total += int64(months) * int64(s.Price)
	}
	return total
}
