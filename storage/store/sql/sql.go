package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/TwiN/gatus/v5/core"
	"github.com/TwiN/gatus/v5/storage/store/common"
	"github.com/TwiN/gatus/v5/storage/store/common/paging"
	"github.com/TwiN/gatus/v5/util"
	"github.com/TwiN/gocache/v2"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

//////////////////////////////////////////////////////////////////////////////////////////////////
// Note that only exported functions in this file may create, commit, or rollback a transaction //
//////////////////////////////////////////////////////////////////////////////////////////////////

const (
	// arraySeparator is the separator used to separate multiple strings in a single column.
	arraySeparator = "|~|"

	// Custom duration constants
	customUptimeRetentionDuration = 30 * 24 * time.Hour // 30 days
	customUptimeCleanupThreshold  = 35 * 24 * time.Hour // 35 days to allow some buffer

	// New custom duration constants
	shortTermUptimeRetentionDuration = 7 * 24 * time.Hour  // 7 days
	mediumTermUptimeRetentionDuration = 14 * 24 * time.Hour // 14 days
	longTermUptimeRetentionDuration = 60 * 24 * time.Hour  // 60 days

	shortTermUptimeCleanupThreshold = 8 * 24 * time.Hour  // 8 days to allow some buffer
	mediumTermUptimeCleanupThreshold = 15 * 24 * time.Hour // 15 days to allow some buffer
	longTermUptimeCleanupThreshold = 65 * 24 * time.Hour  // 65 days to allow some buffer

	uptimeCleanUpThreshold  = 10 * 24 * time.Hour                // Maximum uptime age before triggering a clean up
	eventsCleanUpThreshold  = common.MaximumNumberOfEvents + 10  // Maximum number of events before triggering a clean up
	resultsCleanUpThreshold = common.MaximumNumberOfResults + 10 // Maximum number of results before triggering a clean up

	uptimeRetention = 7 * 24 * time.Hour // Default retention period

	cacheTTL = 10 * time.Minute
)

var (
	// ErrPathNotSpecified is the error returned when the path parameter passed in NewStore is blank
	ErrPathNotSpecified = errors.New("path cannot be empty")

	// ErrDatabaseDriverNotSpecified is the error returned when the driver parameter passed in NewStore is blank
	ErrDatabaseDriverNotSpecified = errors.New("database driver cannot be empty")

	errNoRowsReturned = errors.New("expected a row to be returned, but none was")
)

// Store that leverages a database
type Store struct {
	driver, path string

	db *sql.DB

	// writeThroughCache is a cache used to drastically decrease read latency by pre-emptively
	// caching writes as they happen. If nil, writes are not cached.
	writeThroughCache *gocache.Cache
}

// NewStore initializes the database and creates the schema if it doesn't already exist in the path specified
func NewStore(driver, path string, caching bool) (*Store, error) {
	// ... (rest of the NewStore function remains unchanged)
}

// ... (rest of the functions remain unchanged until the updateEndpointUptime function)

func (s *Store) updateEndpointUptime(tx *sql.Tx, endpointID int64, result *core.Result) error {
	unixTimestampFlooredAtHour := result.Timestamp.Truncate(time.Hour).Unix()
	var successfulExecutions int
	if result.Success {
		successfulExecutions = 1
	}
	_, err := tx.Exec(
		`
			INSERT INTO endpoint_uptimes (endpoint_id, hour_unix_timestamp, total_executions, successful_executions, total_response_time) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT(endpoint_id, hour_unix_timestamp) DO UPDATE SET
				total_executions = excluded.total_executions + endpoint_uptimes.total_executions,
				successful_executions = excluded.successful_executions + endpoint_uptimes.successful_executions,
				total_response_time = excluded.total_response_time + endpoint_uptimes.total_response_time
		`,
		endpointID,
		unixTimestampFlooredAtHour,
		1,
		successfulExecutions,
		result.Duration.Milliseconds(),
	)
	return err
}

// ... (rest of the functions remain unchanged until the deleteOldUptimeEntries function)

func (s *Store) deleteOldUptimeEntries(tx *sql.Tx, endpointID int64, maxAge time.Time) error {
	// Determine the appropriate cleanup threshold based on the maxAge
	var cleanupThreshold time.Duration
	switch {
	case maxAge.Before(time.Now().Add(-longTermUptimeRetentionDuration)):
		cleanupThreshold = longTermUptimeCleanupThreshold
	case maxAge.Before(time.Now().Add(-mediumTermUptimeRetentionDuration)):
		cleanupThreshold = mediumTermUptimeCleanupThreshold
	case maxAge.Before(time.Now().Add(-shortTermUptimeRetentionDuration)):
		cleanupThreshold = shortTermUptimeCleanupThreshold
	default:
		cleanupThreshold = customUptimeCleanupThreshold
	}
	_, err := tx.Exec("DELETE FROM endpoint_uptimes WHERE endpoint_id = $1 AND hour_unix_timestamp < $2", endpointID, maxAge.Add(-cleanupThreshold).Unix())
	return err
}

// ... (rest of the functions remain unchanged)