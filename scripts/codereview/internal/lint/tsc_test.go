package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapTSCSeverity(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected Severity
	}{
		{"error returns high", "error", SeverityHigh},
		{"ERROR uppercase returns high", "ERROR", SeverityHigh},
		{"Error mixed case returns high", "Error", SeverityHigh},
		{"warning returns warning", "warning", SeverityWarning},
		{"WARNING uppercase returns warning", "WARNING", SeverityWarning},
		{"Warning mixed case returns warning", "Warning", SeverityWarning},
		{"info returns info", "info", SeverityInfo},
		{"empty string returns info", "", SeverityInfo},
		{"unknown level returns info", "unknown", SeverityInfo},
		{"note returns info", "note", SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapTSCSeverity(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTSC_Name(t *testing.T) {
	tsc := NewTSC()
	assert.Equal(t, "tsc", tsc.Name())
}

func TestTSC_Language(t *testing.T) {
	tsc := NewTSC()
	assert.Equal(t, LanguageTypeScript, tsc.Language())
}
