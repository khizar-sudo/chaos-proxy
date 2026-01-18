package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/khizar-sudo/chaos-proxy/internal/chaos"
)

func ChaosMiddleware(next http.Handler, engine *chaos.Engine) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decsion := engine.Decide(r)

		// Drop request
		if decsion.Drop {
			fmt.Println("[CHAOS] Dropping request (no response)")
			<-r.Context().Done()
			return
		}

		// Add latency to request
		if decsion.Latency > 0 {
			fmt.Printf("[CHAOS] Injecting latency: %v\n", decsion.Latency)
			select {
			case <-time.After(decsion.Latency):
				// continue after latency
			case <-r.Context().Done():
				fmt.Println("[CHAOS] Request cacnelled during latency")
				return
			}
		}

		// Add error to request
		if decsion.ReturnError {
			fmt.Printf("[CHAOS] Injecting error: %d\n", decsion.ErrorCode)
			http.Error(w, fmt.Sprintf("Chaos injected error %d", decsion.ErrorCode), decsion.ErrorCode)
			return
		}

		// Corrupt the body of the request
		if decsion.Corrupt {
			fmt.Println("[CHAOS] Corrupting response")
			cw := newCorruptionWriter(w)
			next.ServeHTTP(cw, r)
			cw.flush()
			return
		}

		next.ServeHTTP(w, r)
	})
}
