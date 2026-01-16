package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResult(t *testing.T) {
	result := NewResult()

	assert.NotNil(t, result)
	assert.Empty(t, result.Findings)
	assert.Empty(t, result.ToolVersions)
	assert.Empty(t, result.Errors)
	assert.Equal(t, 0, result.Summary.Critical)
	assert.Equal(t, 0, result.Summary.High)
	assert.Equal(t, 0, result.Summary.Warning)
	assert.Equal(t, 0, result.Summary.Info)
	assert.Equal(t, 0, result.Summary.Unknown)
}

func TestResult_AddFinding(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		wantCrit int
		wantHigh int
		wantWarn int
		wantInfo int
		wantUnk  int
	}{
		{"critical", SeverityCritical, 1, 0, 0, 0, 0},
		{"high", SeverityHigh, 0, 1, 0, 0, 0},
		{"warning", SeverityWarning, 0, 0, 1, 0, 0},
		{"info", SeverityInfo, 0, 0, 0, 1, 0},
		{"unknown", Severity("unexpected"), 0, 0, 0, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult()
			finding := Finding{
				Tool:     "test",
				Rule:     "TEST001",
				Severity: tt.severity,
				File:     "test.go",
				Line:     1,
				Column:   1,
				Message:  "test message",
				Category: CategoryBug,
			}

			result.AddFinding(finding)

			assert.Len(t, result.Findings, 1)
			assert.Equal(t, tt.wantCrit, result.Summary.Critical)
			assert.Equal(t, tt.wantHigh, result.Summary.High)
			assert.Equal(t, tt.wantWarn, result.Summary.Warning)
			assert.Equal(t, tt.wantInfo, result.Summary.Info)
			assert.Equal(t, tt.wantUnk, result.Summary.Unknown)
			if tt.wantUnk > 0 {
				assert.Len(t, result.Errors, 1)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestResult_Merge(t *testing.T) {
	result1 := NewResult()
	result1.ToolVersions["tool1"] = "1.0.0"
	result1.AddFinding(Finding{
		Tool: "tool1", Rule: "R001", Severity: SeverityHigh,
		File: "a.go", Line: 1, Message: "issue 1",
	})

	result2 := NewResult()
	result2.ToolVersions["tool2"] = "2.0.0"
	result2.AddFinding(Finding{
		Tool: "tool2", Rule: "R002", Severity: SeverityWarning,
		File: "b.go", Line: 2, Message: "issue 2",
	})
	result2.Errors = append(result2.Errors, "error from tool2")

	result1.Merge(result2)

	assert.Len(t, result1.Findings, 2)
	assert.Equal(t, "1.0.0", result1.ToolVersions["tool1"])
	assert.Equal(t, "2.0.0", result1.ToolVersions["tool2"])
	assert.Equal(t, 1, result1.Summary.High)
	assert.Equal(t, 1, result1.Summary.Warning)
	assert.Len(t, result1.Errors, 1)
}

func TestResult_FilterByFiles(t *testing.T) {
	result := NewResult()
	result.ToolVersions["test"] = "1.0.0"
	result.AddFinding(Finding{
		Tool: "test", Rule: "R001", Severity: SeverityHigh,
		File: "changed.go", Line: 1, Message: "in scope",
	})
	result.AddFinding(Finding{
		Tool: "test", Rule: "R002", Severity: SeverityWarning,
		File: "unchanged.go", Line: 1, Message: "out of scope",
	})
	result.Errors = append(result.Errors, "test error")

	files := map[string]bool{"changed.go": true}
	filtered := result.FilterByFiles(files)

	require.Len(t, filtered.Findings, 1)
	assert.Equal(t, "changed.go", filtered.Findings[0].File)
	assert.Equal(t, 1, filtered.Summary.High)
	assert.Equal(t, 0, filtered.Summary.Warning)
	assert.Equal(t, "1.0.0", filtered.ToolVersions["test"])
	assert.Len(t, filtered.Errors, 1)
}
