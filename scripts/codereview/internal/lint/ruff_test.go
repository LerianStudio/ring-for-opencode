package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapRuffSeverity(t *testing.T) {
	tests := []struct {
		code     string
		expected Severity
	}{
		{"S101", SeverityHigh},    // Security
		{"S501", SeverityHigh},    // Security
		{"E501", SeverityWarning}, // Errors
		{"F401", SeverityWarning}, // Pyflakes
		{"W503", SeverityWarning}, // Warnings
		{"B001", SeverityWarning}, // Bugbear
		{"I001", SeverityInfo},    // Import sorting
		{"D100", SeverityInfo},    // Docstring
		{"N801", SeverityInfo},    // Naming
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := mapRuffSeverity(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapRuffCategory(t *testing.T) {
	tests := []struct {
		code     string
		expected Category
	}{
		{"S101", CategorySecurity},
		{"S501", CategorySecurity},
		{"F401", CategoryBug},
		{"F841", CategoryBug},
		{"E501", CategoryStyle},
		{"E302", CategoryStyle},
		{"W503", CategoryStyle},
		{"W291", CategoryStyle},
		{"B001", CategoryBug},
		{"B007", CategoryBug},
		{"I001", CategoryStyle},
		{"UP001", CategoryDeprecation},
		{"UP035", CategoryDeprecation},
		{"C901", CategoryComplexity},
		{"D100", CategoryOther},
		{"Z999", CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := mapRuffCategory(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRuff_Name(t *testing.T) {
	r := NewRuff()
	assert.Equal(t, "ruff", r.Name())
}

func TestRuff_Language(t *testing.T) {
	r := NewRuff()
	assert.Equal(t, LanguagePython, r.Language())
}
