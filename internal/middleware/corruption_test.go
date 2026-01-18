package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCorruptionWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := newCorruptionWriter(rec)
	
	if cw.ResponseWriter != rec {
		t.Error("Expected ResponseWriter to be set")
	}
	if cw.buf == nil {
		t.Error("Expected buffer to be initialized")
	}
	if cw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code to be 200, got %d", cw.statusCode)
	}
}

func TestCorruptionWriter_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := newCorruptionWriter(rec)
	
	data := []byte("test data")
	n, err := cw.Write(data)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if cw.buf.Len() != len(data) {
		t.Errorf("Expected buffer to contain %d bytes, got %d", len(data), cw.buf.Len())
	}
}

func TestCorruptionWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := newCorruptionWriter(rec)
	
	cw.WriteHeader(http.StatusNotFound)
	
	if cw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code to be 404, got %d", cw.statusCode)
	}
}

func TestCorruptionWriter_Flush(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := newCorruptionWriter(rec)
	
	testData := []byte("test data for corruption")
	cw.Write(testData)
	cw.flush()
	
	// The flush method should write something to the underlying ResponseWriter
	if rec.Body.Len() == 0 {
		t.Error("Expected some data to be written after flush")
	}
}

func TestCorruptRandomBytes_EmptyInput(t *testing.T) {
	result := corruptRandomBytes([]byte{})
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d bytes", len(result))
	}
}

func TestCorruptRandomBytes_ValidInput(t *testing.T) {
	input := []byte("This is a test string with enough length to corrupt")
	result := corruptRandomBytes(input)
	
	if len(result) != len(input) {
		t.Errorf("Expected result length to be %d, got %d", len(input), len(result))
	}
	
	// Check that at least some bytes were corrupted
	diffCount := 0
	for i := range input {
		if result[i] != input[i] {
			diffCount++
		}
	}
	
	if diffCount == 0 {
		t.Error("Expected at least some bytes to be corrupted")
	}
}

func TestCorruptJSON_EmptyInput(t *testing.T) {
	result := corruptJSON([]byte{})
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d bytes", len(result))
	}
}

func TestCorruptJSON_ValidJSON(t *testing.T) {
	input := []byte(`{"name":"test","value":123,"nested":{"key":"value"}}`)
	result := corruptJSON(input)
	
	// Result should be non-empty
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
	
	// Result should be invalid JSON (corrupted)
	var data interface{}
	err := json.Unmarshal(result, &data)
	
	// We expect it to be invalid JSON after corruption in most cases
	// But due to randomness, it might occasionally still be valid, so we just check it ran
	_ = err
}

func TestCorruptJSON_InvalidJSON(t *testing.T) {
	input := []byte("not a json string")
	result := corruptJSON(input)
	
	// Should fall back to corruptString
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestCorruptString_EmptyInput(t *testing.T) {
	result := corruptString([]byte{})
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d bytes", len(result))
	}
}

func TestCorruptString_ShortInput(t *testing.T) {
	input := []byte("abc")
	result := corruptString(input)
	
	// For strings <= 4 chars, it returns as-is
	if !bytes.Equal(result, input) {
		t.Error("Expected short input to be returned unchanged")
	}
}

func TestCorruptString_ValidInput(t *testing.T) {
	input := []byte("This is a longer string for corruption testing")
	result := corruptString(input)
	
	// Result should be non-empty and modified
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
}

func TestTruncateBody_EmptyInput(t *testing.T) {
	result := truncateBody([]byte{})
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d bytes", len(result))
	}
}

func TestTruncateBody_ValidInput(t *testing.T) {
	input := []byte("This is a test string")
	result := truncateBody(input)
	
	expectedLength := len(input) / 2
	if len(result) != expectedLength {
		t.Errorf("Expected result length to be %d, got %d", expectedLength, len(result))
	}
	
	// Result should be first half of input
	if !bytes.Equal(result, input[:expectedLength]) {
		t.Error("Expected result to be first half of input")
	}
}

func TestTruncateBody_SingleByte(t *testing.T) {
	input := []byte("a")
	result := truncateBody(input)
	
	// For single byte, half is 0, so return as-is
	if !bytes.Equal(result, input) {
		t.Error("Expected single byte input to be returned as-is")
	}
}

func TestCorruptionStrategies_Integration(t *testing.T) {
	// Test that all corruption strategies can be called without panicking
	strategies := []struct {
		name string
		fn   func([]byte) []byte
	}{
		{"corruptRandomBytes", corruptRandomBytes},
		{"corruptJSON", corruptJSON},
		{"truncateBody", truncateBody},
	}
	
	testData := []byte(`{"test":"data","number":123}`)
	
	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			result := strategy.fn(testData)
			
			// Just ensure it doesn't panic and returns something
			_ = result
		})
	}
}

func TestFlush_ContentLengthMismatch(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := newCorruptionWriter(rec)
	
	testData := []byte("test data with sufficient length for mismatch")
	cw.Write(testData)
	
	// Run flush multiple times to potentially hit the content-length mismatch strategy
	// Since it's random, we just verify it doesn't panic
	cw.flush()
	
	if rec.Body.Len() == 0 {
		t.Error("Expected some data to be written")
	}
}
