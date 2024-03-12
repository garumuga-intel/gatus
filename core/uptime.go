package core

import (
	"time"
)

// Uptime is the struct that contains the relevant data for calculating the uptime as well as the uptime itself
// and some other statistics
type Uptime struct {
	// HourlyStatistics is a map containing metrics collected (value) for every hourly unix timestamps (key)
	//
	// Used only if the storage type is memory
	HourlyStatistics map[int64]*HourlyUptimeStatistics `json:"-"`
}

// HourlyUptimeStatistics is a struct containing all metrics collected over the course of an hour
type HourlyUptimeStatistics struct {
	TotalExecutions             uint64 // Total number of checks
	SuccessfulExecutions        uint64 // Number of successful executions
	TotalExecutionsResponseTime uint64 // Total response time for all executions in milliseconds
}

// NewUptime creates a new Uptime
func NewUptime() *Uptime {
	return &Uptime{
		HourlyStatistics: make(map[int64]*HourlyUptimeStatistics),
	}
}

// calculateUptimePercentage calculates the uptime percentage over a given period
func (u *Uptime) calculateUptimePercentage(fromTime, toTime int64) float64 {
	var totalExecutions, successfulExecutions uint64

	for timestamp, stats := range u.HourlyStatistics {
		if timestamp >= fromTime && timestamp <= toTime {
			totalExecutions += stats.TotalExecutions
			successfulExecutions += stats.SuccessfulExecutions
		}
	}

	if totalExecutions == 0 {
		return 100.0
	}

	return (float64(successfulExecutions) / float64(totalExecutions)) * 100
}

// UptimeLastMonth calculates the uptime percentage for the last month
func (u *Uptime) UptimeLastMonth() float64 {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	return u.calculateUptimePercentage(lastMonth.Unix(), now.Unix())
}

// UptimeLast90Days calculates the uptime percentage for the last 90 days
func (u *Uptime) UptimeLast90Days() float64 {
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	return u.calculateUptimePercentage(ninetyDaysAgo.Unix(), now.Unix())
}

// UptimeLastYear calculates the uptime percentage for the last year
func (u *Uptime) UptimeLastYear() float64 {
	now := time.Now()
	lastYear := now.AddDate(-1, 0, 0)
	return u.calculateUptimePercentage(lastYear.Unix(), now.Unix())
}