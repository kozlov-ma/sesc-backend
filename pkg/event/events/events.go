// Package events declares reusable keys for event record values.
package events

const (
	// Error key should be used to add unexpected error values to the event.
	Error = "error"

	// PostgresTime is cumulative time spent in postgres to execure the event.
	PostgresTime = "postgres_time"

	// PostgresQueries is cumulative number of postgres queries triggered by the event.
	PostgresQueries = "postgres_queries"
)
