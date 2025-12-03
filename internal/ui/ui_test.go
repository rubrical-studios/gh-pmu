package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestUI_Success(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Success("Operation completed")

	output := buf.String()
	if !strings.Contains(output, SymbolCheck) {
		t.Errorf("Success output should contain check symbol, got: %s", output)
	}
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success output should contain message, got: %s", output)
	}
}

func TestUI_Error(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Error("Something failed")

	output := buf.String()
	if !strings.Contains(output, SymbolCross) {
		t.Errorf("Error output should contain cross symbol, got: %s", output)
	}
	if !strings.Contains(output, "Something failed") {
		t.Errorf("Error output should contain message, got: %s", output)
	}
}

func TestUI_Warning(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Warning("Be careful")

	output := buf.String()
	if !strings.Contains(output, SymbolWarning) {
		t.Errorf("Warning output should contain warning symbol, got: %s", output)
	}
	if !strings.Contains(output, "Be careful") {
		t.Errorf("Warning output should contain message, got: %s", output)
	}
}

func TestUI_Info(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Info("FYI")

	output := buf.String()
	if !strings.Contains(output, SymbolInfo) {
		t.Errorf("Info output should contain info symbol, got: %s", output)
	}
	if !strings.Contains(output, "FYI") {
		t.Errorf("Info output should contain message, got: %s", output)
	}
}

func TestUI_Step(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Step(1, 3, "First step")

	output := buf.String()
	if !strings.Contains(output, "Step 1 of 3") {
		t.Errorf("Step output should contain step indicator, got: %s", output)
	}
	if !strings.Contains(output, "First step") {
		t.Errorf("Step output should contain title, got: %s", output)
	}
}

func TestUI_Header(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Header("Test Header", "Subtitle here")

	output := buf.String()
	if !strings.Contains(output, BoxTopLeft) {
		t.Errorf("Header should contain box characters, got: %s", output)
	}
	if !strings.Contains(output, "Test Header") {
		t.Errorf("Header should contain title, got: %s", output)
	}
	if !strings.Contains(output, "Subtitle here") {
		t.Errorf("Header should contain subtitle, got: %s", output)
	}
}

func TestUI_HeaderWithoutSubtitle(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Header("Title Only", "")

	output := buf.String()
	if !strings.Contains(output, "Title Only") {
		t.Errorf("Header should contain title, got: %s", output)
	}
}

func TestUI_Box(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Box([]string{"Line 1", "Line 2", "Line 3"})

	output := buf.String()
	if !strings.Contains(output, BoxTopLeftAlt) {
		t.Errorf("Box should contain box characters, got: %s", output)
	}
	if !strings.Contains(output, "Line 1") {
		t.Errorf("Box should contain Line 1, got: %s", output)
	}
	if !strings.Contains(output, "Line 2") {
		t.Errorf("Box should contain Line 2, got: %s", output)
	}
	if !strings.Contains(output, "Line 3") {
		t.Errorf("Box should contain Line 3, got: %s", output)
	}
}

func TestUI_BoxEmpty(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	u.Box([]string{})

	output := buf.String()
	if output != "" {
		t.Errorf("Empty box should produce no output, got: %s", output)
	}
}

func TestUI_Menu(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	lines := u.Menu([]string{"Option A", "Option B"}, true)

	if len(lines) != 3 { // 2 options + manual entry
		t.Errorf("Menu should have 3 lines, got: %d", len(lines))
	}

	// Check that lines contain expected content (with ANSI stripped)
	combined := strings.Join(lines, "\n")
	if !strings.Contains(stripANSI(combined), "1.") {
		t.Errorf("Menu should contain option 1, got: %s", combined)
	}
	if !strings.Contains(stripANSI(combined), "Option A") {
		t.Errorf("Menu should contain Option A, got: %s", combined)
	}
	if !strings.Contains(stripANSI(combined), "0.") {
		t.Errorf("Menu should contain manual option 0, got: %s", combined)
	}
}

func TestUI_MenuWithoutManualOption(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	lines := u.Menu([]string{"Option A", "Option B"}, false)

	if len(lines) != 2 {
		t.Errorf("Menu without manual option should have 2 lines, got: %d", len(lines))
	}
}

func TestUI_Prompt(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	prompt := u.Prompt("Enter value", "default")
	if !strings.Contains(stripANSI(prompt), "Enter value") {
		t.Errorf("Prompt should contain label, got: %s", prompt)
	}
	if !strings.Contains(stripANSI(prompt), "default") {
		t.Errorf("Prompt should contain default value, got: %s", prompt)
	}
}

func TestUI_PromptWithoutDefault(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	prompt := u.Prompt("Enter value", "")
	if !strings.Contains(stripANSI(prompt), "Enter value:") {
		t.Errorf("Prompt without default should end with colon, got: %s", prompt)
	}
}

func TestUI_SummaryBox(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	items := map[string]string{
		"Key1": "Value1",
		"Key2": "Value2",
	}
	order := []string{"Key1", "Key2"}

	u.SummaryBox("Summary Title", items, order)

	output := buf.String()
	if !strings.Contains(output, BoxTopLeft) {
		t.Errorf("SummaryBox should contain box characters, got: %s", output)
	}
	if !strings.Contains(output, "Summary Title") {
		t.Errorf("SummaryBox should contain title, got: %s", output)
	}
	if !strings.Contains(output, "Value1") {
		t.Errorf("SummaryBox should contain Value1, got: %s", output)
	}
	if !strings.Contains(output, "Value2") {
		t.Errorf("SummaryBox should contain Value2, got: %s", output)
	}
}

func TestUI_NoColor(t *testing.T) {
	var buf bytes.Buffer
	u := NewWithOptions(&buf, true) // noColor = true

	u.Success("No colors")

	output := buf.String()
	// Should not contain ANSI escape codes
	if strings.Contains(output, "\033[") {
		t.Errorf("NoColor output should not contain ANSI codes, got: %s", output)
	}
	if !strings.Contains(output, SymbolCheck) {
		t.Errorf("NoColor output should still contain symbol, got: %s", output)
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\033[31mred\033[0m", "red"},
		{"\033[1m\033[32mbold green\033[0m", "bold green"},
		{"", ""},
		{"no escape", "no escape"},
	}

	for _, tt := range tests {
		result := stripANSI(tt.input)
		if result != tt.expected {
			t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
	}

	for _, tt := range tests {
		result := max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"plain text", 10},
		{"✓ check", 7},                // checkmark is 1 visible char
		{"\033[32m✓\033[0m check", 7}, // with ANSI codes
		{"", 0},
		{"hello", 5},
		{"日本語", 3},                          // 3 Japanese characters
		{"\033[1m\033[32m✓ Done\033[0m", 6}, // bold green checkmark
	}

	for _, tt := range tests {
		result := visibleWidth(tt.input)
		if result != tt.expected {
			t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
