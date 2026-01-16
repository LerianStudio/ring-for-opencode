package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapBanditSeverity(t *testing.T) {
	tests := []struct {
		name       string
		severity   string
		confidence string
		expected   Severity
	}{
		{"HIGH severity HIGH confidence returns critical", "HIGH", "HIGH", SeverityCritical},
		{"HIGH severity MEDIUM confidence returns high", "HIGH", "MEDIUM", SeverityHigh},
		{"HIGH severity LOW confidence returns high", "HIGH", "LOW", SeverityHigh},
		{"HIGH severity empty confidence returns high", "HIGH", "", SeverityHigh},
		{"MEDIUM severity HIGH confidence returns warning", "MEDIUM", "HIGH", SeverityWarning},
		{"MEDIUM severity MEDIUM confidence returns warning", "MEDIUM", "MEDIUM", SeverityWarning},
		{"MEDIUM severity LOW confidence returns warning", "MEDIUM", "LOW", SeverityWarning},
		{"LOW severity HIGH confidence returns info", "LOW", "HIGH", SeverityInfo},
		{"LOW severity MEDIUM confidence returns info", "LOW", "MEDIUM", SeverityInfo},
		{"LOW severity LOW confidence returns info", "LOW", "LOW", SeverityInfo},
		{"lowercase high high returns critical", "high", "high", SeverityCritical},
		{"mixed case returns correct severity", "High", "Medium", SeverityHigh},
		{"empty severity returns info", "", "HIGH", SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapBanditSeverity(tt.severity, tt.confidence)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBandit_Name(t *testing.T) {
	b := NewBandit()
	assert.Equal(t, "bandit", b.Name())
}

func TestBandit_Language(t *testing.T) {
	b := NewBandit()
	assert.Equal(t, LanguagePython, b.Language())
}
