package lint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapESLintSeverity(t *testing.T) {
	tests := []struct {
		input    int
		expected Severity
	}{
		{2, SeverityHigh},
		{1, SeverityWarning},
		{0, SeverityInfo},
		{99, SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("severity_%d", tt.input), func(t *testing.T) {
			result := mapESLintSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapESLintCategory(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected Category
	}{
		{"@typescript-eslint/no-unused-vars", CategoryType},
		{"@typescript-eslint/explicit-function-return-type", CategoryType},
		{"security/detect-object-injection", CategorySecurity},
		{"no-unused-vars", CategoryUnused},
		{"no-unused-expressions", CategoryUnused},
		{"import/order", CategoryStyle},
		{"import/no-unresolved", CategoryStyle},
		{"react/jsx-uses-react", CategoryStyle},
		{"react-hooks/rules-of-hooks", CategoryStyle},
		{"parse-error", CategoryBug},
		{"semi", CategoryStyle},
		{"unknown-rule", CategoryStyle},
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			result := mapESLintCategory(tt.ruleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestESLint_Name(t *testing.T) {
	e := NewESLint()
	assert.Equal(t, "eslint", e.Name())
}

func TestESLint_Language(t *testing.T) {
	e := NewESLint()
	assert.Equal(t, LanguageTypeScript, e.Language())
}
