package chaos

import (
	"net/http"
	"testing"
	"time"
)

// TestDecide_Drop tests that drop behavior is exclusive and terminal
func TestDecide_Drop(t *testing.T) {
	engine := NewEngine(ChaosConfig{
		DropRate: 100, // Always drop

		// These should be ignored when drop is triggered
		ErrorRate:   100,
		LatencyMin:  1 * time.Second,
		LatencyMax:  2 * time.Second,
		CorruptRate: 100,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if !decision.Drop {
		t.Error("Expected Drop to be true")
	}
	if decision.ReturnError {
		t.Error("Expected ReturnError to be false when dropping")
	}
	if decision.Latency != 0 {
		t.Error("Expected no latency when dropping")
	}
	if decision.Corrupt {
		t.Error("Expected Corrupt to be false when dropping")
	}
}

// TestDecide_NoChaosBehavior tests that no chaos is applied when rates are 0
func TestDecide_NoChaosBehavior(t *testing.T) {
	// Empty struct has all 0 values
	engine := NewEngine(ChaosConfig{})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
	if decision.ReturnError {
		t.Error("Expected ReturnError to be false")
	}
	if decision.Latency != 0 {
		t.Error("Expected no latency")
	}
	if decision.Corrupt {
		t.Error("Expected Corrupt to be false")
	}
}

// TestDecide_ErrorOnly tests error behavior without other chaos
func TestDecide_ErrorOnly(t *testing.T) {
	tests := []struct {
		name              string
		errorCode         int
		expectedErrorCode int
	}{
		{
			name:              "custom error code",
			errorCode:         503,
			expectedErrorCode: 503,
		},
		{
			name:              "default error code",
			errorCode:         0,
			expectedErrorCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(ChaosConfig{
				ErrorRate: 100,
				ErrorCode: tt.errorCode,
			})

			req, _ := http.NewRequest("GET", "http://example.com", nil)
			decision := engine.Decide(req)

			if !decision.ReturnError {
				t.Error("Expected ReturnError to be true")
			}
			if decision.ErrorCode != tt.expectedErrorCode {
				t.Errorf("Expected ErrorCode to be %d, got %d", tt.expectedErrorCode, decision.ErrorCode)
			}
			if decision.Drop {
				t.Error("Expected Drop to be false")
			}
		})
	}
}

// TestDecide_FixedLatency tests fixed latency application
func TestDecide_FixedLatency(t *testing.T) {
	expectedLatency := 500 * time.Millisecond
	engine := NewEngine(ChaosConfig{
		Latency: expectedLatency,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if decision.Latency != expectedLatency {
		t.Errorf("Expected latency to be %v, got %v", expectedLatency, decision.Latency)
	}
}

// TestDecide_RandomLatency tests random latency range
func TestDecide_RandomLatency(t *testing.T) {
	minLatency := 100 * time.Millisecond
	maxLatency := 500 * time.Millisecond

	engine := NewEngine(ChaosConfig{
		LatencyMin: minLatency,
		LatencyMax: maxLatency,
	})

	// Run multiple times to test randomness
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		decision := engine.Decide(req)

		if decision.Latency < minLatency {
			t.Errorf("Latency %v is less than minimum %v", decision.Latency, minLatency)
		}
		if decision.Latency > maxLatency {
			t.Errorf("Latency %v is greater than maximum %v", decision.Latency, maxLatency)
		}
	}
}

// TestDecide_CorruptOnly tests corruption behavior
func TestDecide_CorruptOnly(t *testing.T) {
	engine := NewEngine(ChaosConfig{
		CorruptRate: 100,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if !decision.Corrupt {
		t.Error("Expected Corrupt to be true")
	}
	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
	if decision.ReturnError {
		t.Error("Expected ReturnError to be false")
	}
}

// TestDecide_ErrorWithLatency tests combination of error and latency
func TestDecide_ErrorWithLatency(t *testing.T) {
	expectedLatency := 2 * time.Second
	expectedErrorCode := 504

	engine := NewEngine(ChaosConfig{
		ErrorRate: 100,
		ErrorCode: expectedErrorCode,
		Latency:   expectedLatency,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if !decision.ReturnError {
		t.Error("Expected ReturnError to be true")
	}
	if decision.ErrorCode != expectedErrorCode {
		t.Errorf("Expected ErrorCode to be %d, got %d", expectedErrorCode, decision.ErrorCode)
	}
	if decision.Latency != expectedLatency {
		t.Errorf("Expected latency to be %v, got %v", expectedLatency, decision.Latency)
	}
	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
	if decision.Corrupt {
		t.Error("Expected Corrupt to be false")
	}
}

// TestDecide_ErrorWithCorrupt tests combination of error and corruption
func TestDecide_ErrorWithCorrupt(t *testing.T) {
	errorCode := 503
	engine := NewEngine(ChaosConfig{
		ErrorRate:   100,
		ErrorCode:   errorCode,
		CorruptRate: 100,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if !decision.ReturnError {
		t.Error("Expected ReturnError to be true")
	}
	if decision.ErrorCode != errorCode {
		t.Errorf("Expected error code to be %v, got %v", errorCode, decision.ErrorCode)
	}
	if !decision.Corrupt {
		t.Error("Expected Corrupt to be true")
	}
	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
	if decision.Latency > 0 {
		t.Error("Expected Latency to be 0")
	}
}

// TestDecide_LatencyWithCorrupt tests combination of latency and corruption
func TestDecide_LatencyWithCorrupt(t *testing.T) {
	expectedLatency := 1 * time.Second

	engine := NewEngine(ChaosConfig{
		Latency:     expectedLatency,
		CorruptRate: 100,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if decision.Latency != expectedLatency {
		t.Errorf("Expected latency to be %v, got %v", expectedLatency, decision.Latency)
	}
	if !decision.Corrupt {
		t.Error("Expected Corrupt to be true")
	}
	if decision.ReturnError {
		t.Error("Expected ReturnError to be false")
	}
	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
}

// TestDecide_AllCombined tests the "perfect storm" - error, latency, and corruption together
func TestDecide_AllCombined(t *testing.T) {
	expectedLatency := 3 * time.Second
	expectedErrorCode := 500

	engine := NewEngine(ChaosConfig{
		ErrorRate:   100,
		ErrorCode:   expectedErrorCode,
		Latency:     expectedLatency,
		CorruptRate: 100,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if !decision.ReturnError {
		t.Error("Expected ReturnError to be true")
	}
	if decision.ErrorCode != expectedErrorCode {
		t.Errorf("Expected ErrorCode to be %d, got %d", expectedErrorCode, decision.ErrorCode)
	}
	if decision.Latency != expectedLatency {
		t.Errorf("Expected latency to be %v, got %v", expectedLatency, decision.Latency)
	}
	if !decision.Corrupt {
		t.Error("Expected Corrupt to be true")
	}
	if decision.Drop {
		t.Error("Expected Drop to be false")
	}
}

// TestShouldApply_Rates tests the probability logic
func TestShouldApply_Rates(t *testing.T) {
	tests := []struct {
		name            string
		rate            float64
		expectedResult  bool
		exactMatch      bool // For 0 and 100, we expect exact behavior
		expectedMinRate float64
		expectedMaxRate float64
	}{
		{
			name:           "rate 0 never applies",
			rate:           0,
			expectedResult: false,
			exactMatch:     true,
		},
		{
			name:           "rate 100 always applies",
			rate:           100,
			expectedResult: true,
			exactMatch:     true,
		},
		{
			name:            "rate 50 applies approximately half the time",
			rate:            50,
			exactMatch:      false,
			expectedMinRate: 40, // Allow 40-60% range
			expectedMaxRate: 60,
		},
		{
			name:            "rate 10 applies approximately 10% of the time",
			rate:            10,
			exactMatch:      false,
			expectedMinRate: 5, // Allow 5-15% range
			expectedMaxRate: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(ChaosConfig{})

			if tt.exactMatch {
				// For 0 and 100, test exact behavior
				for i := 0; i < 10; i++ {
					result := engine.shouldApply(tt.rate)
					if result != tt.expectedResult {
						t.Errorf("Expected shouldApply(%v) to be %v, got %v", tt.rate, tt.expectedResult, result)
					}
				}
			} else {
				// For probabilistic rates, test over many iterations
				iterations := 1000
				trueCount := 0

				for i := 0; i < iterations; i++ {
					if engine.shouldApply(tt.rate) {
						trueCount++
					}
				}

				actualRate := float64(trueCount) / float64(iterations) * 100

				if actualRate < tt.expectedMinRate || actualRate > tt.expectedMaxRate {
					t.Errorf("Expected rate to be between %v%% and %v%%, got %v%%",
						tt.expectedMinRate, tt.expectedMaxRate, actualRate)
				}
			}
		})
	}
}

// TestDecide_FixedLatencyTakesPrecedence tests that fixed latency takes precedence over random
func TestDecide_FixedLatencyTakesPrecedence(t *testing.T) {
	fixedLatency := 1 * time.Second

	engine := NewEngine(ChaosConfig{
		Latency:    fixedLatency,
		LatencyMin: 100 * time.Millisecond,
		LatencyMax: 200 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	decision := engine.Decide(req)

	if decision.Latency != fixedLatency {
		t.Errorf("Expected fixed latency %v to take precedence, got %v", fixedLatency, decision.Latency)
	}
}
