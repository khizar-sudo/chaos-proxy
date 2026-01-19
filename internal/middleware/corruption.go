package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand" // #nosec G404 - math/rand is sufficient for chaos testing, cryptographic randomness not required
	"net/http"
	"strings"
)

type corruptingWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func newCorruptionWriter(w http.ResponseWriter) *corruptingWriter {
	return &corruptingWriter{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (cw *corruptingWriter) Write(b []byte) (int, error) {
	return cw.buf.Write(b)
}

func (cw *corruptingWriter) WriteHeader(statusCode int) {
	cw.statusCode = statusCode
}

func (cw *corruptingWriter) flush() {
	body := cw.buf.Bytes()

	// Randomly select corruption strategy
	strategy := rand.Intn(4) // #nosec G404 - chaos testing doesn't need crypto rand
	var corrupted []byte
	var strategyName string

	switch strategy {
	case 0:
		corrupted = corruptRandomBytes(body)
		strategyName = "Random Byte Corruption"
	case 1:
		corrupted = corruptJSON(body)
		strategyName = "JSON Corruption"
	case 2:
		corrupted = truncateBody(body)
		strategyName = "Truncation"
	case 3:
		corrupted = body
		strategyName = "Content-Length Mismatch"

		wrongLength := len(body) / 2
		if wrongLength == 0 {
			wrongLength = 10
		}
		cw.ResponseWriter.Header().Set("Content-Length", fmt.Sprintf("%d", wrongLength))
		fmt.Printf("[CHAOS] Set Content-Length to %d (actual: %d)\n", wrongLength, len(body))
	default:
		corrupted = body
		strategyName = "None"
	}

	fmt.Printf("[CHAOS] Strategy: %s | %d bytes -> %d bytes\n", strategyName, len(body), len(corrupted))

	cw.ResponseWriter.WriteHeader(cw.statusCode)
	if _, err := cw.ResponseWriter.Write(corrupted); err != nil {
		fmt.Printf("[CHAOS] Error writing corrupted response: %v\n", err)
	}
}

// Strategy 1: Random Byte Corruption
func corruptRandomBytes(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	corrupted := make([]byte, len(body))
	copy(corrupted, body)

	// Corrupt 5-20% of bytes
	corruptionRate := 0.05 + rand.Float64()*0.15 // #nosec G404 - chaos testing doesn't need crypto rand
	numCorruptions := int(float64(len(body)) * corruptionRate)
	if numCorruptions == 0 && len(body) > 0 {
		numCorruptions = 1
	}

	for i := 0; i < numCorruptions; i++ {
		pos := rand.Intn(len(corrupted))      // #nosec G404 - chaos testing doesn't need crypto rand
		corrupted[pos] = byte(rand.Intn(256)) // #nosec G404 - chaos testing doesn't need crypto rand
	}

	return corrupted
}

// Strategy 2: JSON-Specific Corruption
func corruptJSON(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return corruptString(body)
	}

	bodyStr := string(body)

	// Choose a JSON corruption method
	method := rand.Intn(5) // #nosec G404 - chaos testing doesn't need crypto rand
	switch method {
	case 0:
		// Remove random closing bracket/brace
		bodyStr = strings.Replace(bodyStr, "}", "", 1)
	case 1:
		// Add extra comma
		if idx := strings.Index(bodyStr, ","); idx > 0 {
			bodyStr = bodyStr[:idx+1] + "," + bodyStr[idx+1:]
		}
	case 2:
		// Remove random quote
		if idx := strings.Index(bodyStr, "\""); idx > 0 {
			bodyStr = bodyStr[:idx] + bodyStr[idx+1:]
		}
	case 3:
		// Replace a colon with something else
		bodyStr = strings.Replace(bodyStr, ":", "=", 1)
	case 4:
		// Inject random characters in the middle
		if len(bodyStr) > 2 {
			mid := len(bodyStr) / 2
			bodyStr = bodyStr[:mid] + "XXX" + bodyStr[mid:]
		}
	}

	return []byte(bodyStr)
}

// Helper: Corrupt string content
func corruptString(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	bodyStr := string(body)
	if len(bodyStr) > 4 {
		mid := len(bodyStr) / 2
		bodyStr = bodyStr[:mid] + "���" + bodyStr[mid+3:]
	}
	return []byte(bodyStr)
}

// Strategy 3: Truncation
func truncateBody(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	half := len(body) / 2
	if half > 0 {
		return body[:half]
	}
	return body
}
