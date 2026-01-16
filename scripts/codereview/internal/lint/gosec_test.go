package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapGosecSeverity(t *testing.T) {
	tests := []struct {
		name       string
		severity   string
		confidence string
		expected   Severity
	}{
		{"HIGH severity HIGH confidence returns critical", "HIGH", "HIGH", SeverityCritical},
		{"HIGH severity MEDIUM confidence returns high", "HIGH", "MEDIUM", SeverityHigh},
		{"HIGH severity LOW confidence returns high", "HIGH", "LOW", SeverityHigh},
		{"MEDIUM severity HIGH confidence returns warning", "MEDIUM", "HIGH", SeverityWarning},
		{"MEDIUM severity MEDIUM confidence returns warning", "MEDIUM", "MEDIUM", SeverityWarning},
		{"LOW severity HIGH confidence returns info", "LOW", "HIGH", SeverityInfo},
		{"LOW severity LOW confidence returns info", "LOW", "LOW", SeverityInfo},
		{"lowercase high high returns critical", "high", "high", SeverityCritical},
		{"mixed case returns correct severity", "High", "Medium", SeverityHigh},
		{"empty severity returns info", "", "HIGH", SeverityInfo},
		{"empty confidence returns appropriate severity", "HIGH", "", SeverityHigh},
		{"empty severity and confidence returns info", "", "", SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapGosecSeverity(tt.severity, tt.confidence)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGosec_Name(t *testing.T) {
	g := NewGosec()
	assert.Equal(t, "gosec", g.Name())
}

func TestGosec_Language(t *testing.T) {
	g := NewGosec()
	assert.Equal(t, LanguageGo, g.Language())
}
