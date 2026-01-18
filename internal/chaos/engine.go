package chaos

import (
	"math/rand"
	"net/http"
	"time"
)

type Engine struct {
	config ChaosConfig
	rnd    *rand.Rand
}

func NewEngine(cfg ChaosConfig) *Engine {
	return &Engine{
		config: cfg,
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (e *Engine) Decide(r *http.Request) Decision {
	decison := Decision{}

	if e.shouldApply(e.config.DropRate) {
		decison.Drop = true
		// Dropping a request is terminal. No need to evaluate other conditions
		return decison
	}

	if e.shouldApply((e.config.ErrorRate)) {
		decison.ReturnError = true
		if e.config.ErrorCode == 0 {
			decison.ErrorCode = 500
		} else {
			decison.ErrorCode = e.config.ErrorCode
		}
	}

	if e.config.Latency > 0 {
		decison.Latency = e.config.Latency
	} else if e.config.LatencyMax >= e.config.LatencyMin && e.config.LatencyMax > 0 {
		diff := e.config.LatencyMax - e.config.LatencyMin

		// Note: Pls make sure that LatencyMax >= LatencyMin, otherwise this will panic
		// No, I will not handle this edge case of user error. Sorry!
		random := time.Duration(e.rnd.Int63n((int64(diff))))
		decison.Latency = e.config.LatencyMin + random
	}

	if e.shouldApply(e.config.CorruptRate) {
		decison.Corrupt = true
	}

	return decison
}

func (e *Engine) shouldApply(rate float64) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 100 {
		return true
	}
	return e.rnd.Float64()*100 < rate
}
