package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapMypySeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		expected Severity
	}{
		{"error returns high", "error", SeverityHigh},
		{"ERROR uppercase returns high", "ERROR", SeverityHigh},
		{"Error mixed case returns high", "Error", SeverityHigh},
		{"warning returns warning", "warning", SeverityWarning},
		{"WARNING uppercase returns warning", "WARNING", SeverityWarning},
		{"note returns info", "note", SeverityInfo},
		{"NOTE uppercase returns info", "NOTE", SeverityInfo},
		{"empty string returns warning", "", SeverityWarning},
		{"unknown severity returns warning", "unknown", SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMypySeverity(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMypy_Name(t *testing.T) {
	m := NewMypy()
	assert.Equal(t, "mypy", m.Name())
}

func TestMypy_Language(t *testing.T) {
	m := NewMypy()
	assert.Equal(t, LanguagePython, m.Language())
}
