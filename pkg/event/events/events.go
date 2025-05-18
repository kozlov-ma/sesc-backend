// Package events declares reusable keys for event record values.
package events

const (
	// Error key should be used to add unexpected error values to the event.
	Error = "error"

	// PostgresTime is cumulative time spent in postgres to execure the event.
	PostgresTime = "postgres_time"

	// PostgresQueries is cumulative number of postgres queries triggered by the event.
	PostgresQueries = "postgres_queries"

	// Critical key should be set to true to signal to the sinks that this event should be stored and marked as an error.
	Critical = "critical"
)
