package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Metrics holds runtime performance statistics
type Metrics struct {
	StartTime            time.Time `json:"start_time"`
	UptimeSeconds        int64     `json:"uptime_seconds"`
	TotalRequests        uint64    `json:"total_requests"`
	SuccessCount         uint64    `json:"success_count"`
	FailureCount         uint64    `json:"failure_count"`
	AndroidCount         uint64    `json:"android_count"`
	IOSCount             uint64    `json:"ios_count"`
	WebCount             uint64    `json:"web_count"`
}

var globalMetrics = &Metrics{
	StartTime: time.Now(),
}

// RecordRequest increments request counters based on platform and outcome
func RecordRequest(platform string, success bool) {
	atomic.AddUint64(&globalMetrics.TotalRequests, 1)

	if success {
		atomic.AddUint64(&globalMetrics.SuccessCount, 1)
	} else {
		atomic.AddUint64(&globalMetrics.FailureCount, 1)
	}

	switch platform {
	case "android", "android_rn":
		atomic.AddUint64(&globalMetrics.AndroidCount, 1)
	case "apple", "apple_rn":
		atomic.AddUint64(&globalMetrics.IOSCount, 1)
	case "web", "desktop":
		atomic.AddUint64(&globalMetrics.WebCount, 1)
	}
}

// Handler returns the metrics JSON HTTP handler
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := Metrics{
			StartTime:     globalMetrics.StartTime,
			UptimeSeconds: int64(time.Since(globalMetrics.StartTime).Seconds()),
			TotalRequests: atomic.LoadUint64(&globalMetrics.TotalRequests),
			SuccessCount:  atomic.LoadUint64(&globalMetrics.SuccessCount),
			FailureCount:  atomic.LoadUint64(&globalMetrics.FailureCount),
			AndroidCount:  atomic.LoadUint64(&globalMetrics.AndroidCount),
			IOSCount:      atomic.LoadUint64(&globalMetrics.IOSCount),
			WebCount:      atomic.LoadUint64(&globalMetrics.WebCount),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(m)
	}
}
