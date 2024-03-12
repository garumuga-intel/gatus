package memory

import (
	"time"

	"github.com/TwiN/gatus/v5/core"
)

const (
	numberOfHoursInTenDays = 10 * 24
	sevenDays              = 7 * 24 * time.Hour
	// Custom durations
	thirtyDays = 30 * 24 * time.Hour
	ninetyDays = 90 * 24 * time.Hour
	// Additional custom durations
	sixtyDays  = 60 * 24 * time.Hour
	oneHundredTwentyDays = 120 * 24 * time.Hour
)

// processUptimeAfterResult processes the result by extracting the relevant from the result and recalculating the uptime
// if necessary
func processUptimeAfterResult(uptime *core.Uptime, result *core.Result) {
	if uptime.HourlyStatistics == nil {
		uptime.HourlyStatistics = make(map[int64]*core.HourlyUptimeStatistics)
	}
	unixTimestampFlooredAtHour := result.Timestamp.Truncate(time.Hour).Unix()
	hourlyStats, _ := uptime.HourlyStatistics[unixTimestampFlooredAtHour]
	if hourlyStats == nil {
		hourlyStats = &core.HourlyUptimeStatistics{}
		uptime.HourlyStatistics[unixTimestampFlooredAtHour] = hourlyStats
	}
	if result.Success {
		hourlyStats.SuccessfulExecutions++
	}
	hourlyStats.TotalExecutions++
	hourlyStats.TotalExecutionsResponseTime += uint64(result.Duration.Milliseconds())

	// Clean up mechanism for custom durations
	cleanupOldStatistics(uptime, thirtyDays)
	cleanupOldStatistics(uptime, sixtyDays)
	cleanupOldStatistics(uptime, ninetyDays)
	cleanupOldStatistics(uptime, oneHundredTwentyDays)

	// Update uptime percentage calculations
	updateUptimePercentages(uptime)
}

// cleanupOldStatistics removes statistics older than the specified duration
func cleanupOldStatistics(uptime *core.Uptime, duration time.Duration) {
	threshold := time.Now().Add(-duration).Unix()
	for timestamp := range uptime.HourlyStatistics {
		if timestamp < threshold {
			delete(uptime.HourlyStatistics, timestamp)
		}
	}
}

// updateUptimePercentages recalculates the uptime percentages for different durations
func updateUptimePercentages(uptime *core.Uptime) {
	now := time.Now()
	uptime.SevenDayPercentage = calculateUptimePercentage(uptime, now.Add(-sevenDays))
	uptime.ThirtyDayPercentage = calculateUptimePercentage(uptime, now.Add(-thirtyDays))
	uptime.SixtyDayPercentage = calculateUptimePercentage(uptime, now.Add(-sixtyDays))
	uptime.NinetyDayPercentage = calculateUptimePercentage(uptime, now.Add(-ninetyDays))
	uptime.OneHundredTwentyDayPercentage = calculateUptimePercentage(uptime, now.Add(-oneHundredTwentyDays))
}

// calculateUptimePercentage calculates the uptime percentage for the duration up to the specified end time
func calculateUptimePercentage(uptime *core.Uptime, endTime time.Time) float64 {
	var successfulExecutions, totalExecutions uint64
	endTimestamp := endTime.Unix()
	for timestamp, stats := range uptime.HourlyStatistics {
		if timestamp >= endTimestamp {
			successfulExecutions += stats.SuccessfulExecutions
			totalExecutions += stats.TotalExecutions
		}
	}
	if totalExecutions == 0 {
		return 100
	}
	return float64(successfulExecutions) / float64(totalExecutions) * 100
}