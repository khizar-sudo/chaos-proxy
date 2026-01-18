package chaos

import "time"

// Final decision for the request
type Decision struct {
	Drop        bool
	ReturnError bool
	ErrorCode   int
	Latency     time.Duration
	Corrupt     bool
}

// Values for each error which will give the decision
type ChaosConfig struct {
	ErrorRate   float64       //0-100 percentage
	ErrorCode   int           // HTTP status code to return
	DropRate    float64       //0-100 percentage
	Latency     time.Duration // Fixed latency
	LatencyMin  time.Duration // Min random latency
	LatencyMax  time.Duration // Max random latency
	CorruptRate float64       //0-100 percentage
}
