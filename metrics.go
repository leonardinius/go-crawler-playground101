package main

import (
	"fmt"
	"sync/atomic"
)

// Metrics holds the metrics for the crawler.
type Metrics struct {
	inflightRequests    atomic.Int64
	maxInflightRequests atomic.Int64
	errors              atomic.Int64
	reportedTotal       atomic.Int64
	visitedTotal        atomic.Int64
}

// NewMetrics creates a new metrics object.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// ClearMetrics resets all metrics to zero.
func (m *Metrics) ClearMetrics() {
	m.inflightRequests.Store(0)
	m.maxInflightRequests.Store(0)
	m.errors.Store(0)
	m.reportedTotal.Store(0)
	m.visitedTotal.Store(0)
}

// IncInflightRequests increments the number of inflight requests.
func (m *Metrics) IncInflightRequests() {
	val := m.inflightRequests.Add(1)
	if max := m.maxInflightRequests.Load(); max < val {
		m.maxInflightRequests.CompareAndSwap(max, val)
	}
}

// DecInflightRequests decrements the number of inflight requests.
func (m *Metrics) DecInflightRequests() {
	m.inflightRequests.Add(-1)
}

// IncReportedTotal increments the number of reported URLs.
func (m *Metrics) IncReportedTotal(delta int64) {
	m.reportedTotal.Add(delta)
}

// IncVisitedTotal increments the number of visited URLs.(m *metrics)IncVisitedTotal() {
func (m *Metrics) IncVisitedTotal() {
	m.visitedTotal.Add(1)
}

// IncErrors increments the number of errors.
func (m *Metrics) IncErrors() {
	m.errors.Add(1)
}

// Summary returns a string with the current metrics.
func (m *Metrics) Summary() string {
	return fmt.Sprintf(
		"reported_total=%d, visited=%d err=%d, inflight=%d, max=%d",
		m.reportedTotal.Load(),
		m.visitedTotal.Load(),
		m.errors.Load(),
		m.inflightRequests.Load(),
		m.maxInflightRequests.Load())
}
