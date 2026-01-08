package main

import (
	"strings"
	"testing"
)

// ===== Markdown Rendering Tests =====

func TestRenderMarkdownHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "H1 header",
			input: "# Hello World",
			contains: []string{
				`<h1 id="hello-world"`,
				`Hello World`,
				`class="header-anchor"`,
			},
		},
		{
			name:  "H2 header",
			input: "## Section Title",
			contains: []string{
				`<h2 id="section-title"`,
				`Section Title`,
			},
		},
		{
			name:  "Multiple headers generate TOC",
			input: "# One\n## Two\n## Three\n### Four",
			contains: []string{
				`class="toc"`,
				`Table of Contents`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderMarkdown(tt.input, "")
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("renderMarkdown() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderMarkdownInlineFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"Bold with asterisks", "**bold**", "<strong>bold</strong>"},
		{"Bold with underscores", "__bold__", "<strong>bold</strong>"},
		{"Italic", "*italic*", "<em>italic</em>"},
		{"Strikethrough", "~~deleted~~", "<del>deleted</del>"},
		{"Inline code", "`code`", "<code>code</code>"},
		{"Highlight", "==highlight==", "<mark>highlight</mark>"},
		{"Link", "[text](http://example.com)", `<a href="http://example.com">text</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderMarkdown(tt.input, "")
			if !strings.Contains(result, tt.contains) {
				t.Errorf("renderMarkdown(%q) = %q, want to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

func TestRenderMarkdownCodeBlocks(t *testing.T) {
	input := "```python\nprint('hello')\n```"
	result := renderMarkdown(input, "")

	if !strings.Contains(result, `class="language-python"`) {
		t.Error("Code block should have language class")
	}
	if !strings.Contains(result, "print(") {
		t.Error("Code block should contain code content")
	}
	if !strings.Contains(result, "copy-btn") {
		t.Error("Code block should have copy button")
	}
}

func TestRenderMarkdownMermaid(t *testing.T) {
	input := "```mermaid\ngraph TD\n    A-->B\n```"
	result := renderMarkdown(input, "")

	if !strings.Contains(result, `class="mermaid"`) {
		t.Error("Mermaid block should have mermaid class")
	}
	if !strings.Contains(result, "graph TD") {
		t.Error("Mermaid block should contain diagram content")
	}
}

func TestRenderMarkdownPlantUML(t *testing.T) {
	input := "```plantuml\n@startuml\nAlice -> Bob\n@enduml\n```"
	result := renderMarkdown(input, "")

	if !strings.Contains(result, `class="plantuml"`) {
		t.Error("PlantUML block should have plantuml class")
	}
	if !strings.Contains(result, "plantuml.com/plantuml/svg/") {
		t.Error("PlantUML block should contain server URL")
	}
}

func TestRenderMarkdownLists(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"Unordered list", "- item 1\n- item 2", "<ul>"},
		{"Ordered list", "1. first\n2. second", "<ol>"},
		{"Task list checked", "- [x] done", `<span class="checkbox checked">`},
		{"Task list unchecked", "- [ ] todo", `<span class="checkbox unchecked">`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderMarkdown(tt.input, "")
			if !strings.Contains(result, tt.contains) {
				t.Errorf("renderMarkdown(%q) missing %q", tt.input, tt.contains)
			}
		})
	}
}

func TestRenderMarkdownBlockquote(t *testing.T) {
	input := "> This is a quote"
	result := renderMarkdown(input, "")

	if !strings.Contains(result, "<blockquote>") {
		t.Error("Should contain blockquote tag")
	}
	if !strings.Contains(result, "This is a quote") {
		t.Error("Should contain quote content")
	}
}

func TestRenderMarkdownTable(t *testing.T) {
	input := "| Col1 | Col2 |\n|------|------|\n| A    | B    |"
	result := renderMarkdown(input, "")

	if !strings.Contains(result, "<table>") {
		t.Error("Should contain table tag")
	}
	if !strings.Contains(result, "<th>") {
		t.Error("Should contain header cells")
	}
	if !strings.Contains(result, "<td>") {
		t.Error("Should contain data cells")
	}
}

func TestRenderMarkdownFootnotes(t *testing.T) {
	input := "Text with footnote[^1].\n\n[^1]: Footnote content."
	result := renderMarkdown(input, "")

	if !strings.Contains(result, `class="footnote-ref"`) {
		t.Error("Should contain footnote reference")
	}
	if !strings.Contains(result, `class="footnotes"`) {
		t.Error("Should contain footnotes section")
	}
	if !strings.Contains(result, "Footnote content") {
		t.Error("Should contain footnote content")
	}
}

func TestRenderMarkdownMath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"Inline math", "$E = mc^2$", `class="math-inline"`},
		{"Block math", "$$\nx^2\n$$", `class="math-block"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderMarkdown(tt.input, "")
			if !strings.Contains(result, tt.contains) {
				t.Errorf("renderMarkdown(%q) missing %q", tt.input, tt.contains)
			}
		})
	}
}

func TestRenderMarkdownHorizontalRule(t *testing.T) {
	inputs := []string{"---", "***", "___"}
	for _, input := range inputs {
		result := renderMarkdown(input, "")
		if !strings.Contains(result, "<hr>") {
			t.Errorf("renderMarkdown(%q) should contain <hr>", input)
		}
	}
}

// ===== JSON Rendering Tests =====

func TestRenderJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "Simple object",
			input: `{"name": "test", "value": 123}`,
			contains: []string{
				`class="json-key"`,
				`class="json-string"`,
				`class="json-number"`,
			},
		},
		{
			name:  "Array",
			input: `[1, 2, 3]`,
			contains: []string{
				`class="json-bracket"`,
				`class="json-number"`,
			},
		},
		{
			name:  "Boolean and null",
			input: `{"active": true, "data": null}`,
			contains: []string{
				`class="json-boolean"`,
				`class="json-null"`,
			},
		},
		{
			name:     "Invalid JSON",
			input:    `{invalid}`,
			contains: []string{`class="error"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderJSON(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("renderJSON() missing %q", want)
				}
			}
		})
	}
}

// ===== YAML Rendering Tests =====

func TestRenderYAML(t *testing.T) {
	input := "server:\n  host: localhost\n  port: 8080"
	result := renderYAML(input)

	if !strings.Contains(result, `class="language-yaml"`) {
		t.Error("Should have YAML language class")
	}
	if !strings.Contains(result, "yaml-toolbar") {
		t.Error("Should have YAML toolbar")
	}
	if !strings.Contains(result, "server:") {
		t.Error("Should contain YAML content")
	}
}

// ===== TOML Rendering Tests =====

func TestRenderTOML(t *testing.T) {
	input := "[server]\nhost = \"localhost\"\nport = 8080"
	result := renderTOML(input)

	if !strings.Contains(result, `class="language-toml"`) {
		t.Error("Should have TOML language class")
	}
	if !strings.Contains(result, "toml-toolbar") {
		t.Error("Should have TOML toolbar")
	}
}

// ===== CSV Rendering Tests =====

func TestRenderCSV(t *testing.T) {
	input := "Name,Age,City\nJohn,30,NYC\nJane,25,LA"
	result := renderCSV(input)

	if !strings.Contains(result, `class="csv-table"`) {
		t.Error("Should have CSV table class")
	}
	if !strings.Contains(result, "<th>Name</th>") {
		t.Error("Should have header cells")
	}
	if !strings.Contains(result, "<td>John</td>") {
		t.Error("Should have data cells")
	}
}

func TestParseCSVLine(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "Simple",
			input:  "a,b,c",
			expect: []string{"a", "b", "c"},
		},
		{
			name:   "Quoted fields",
			input:  `"hello, world",test,"quoted"`,
			expect: []string{"hello, world", "test", "quoted"},
		},
		{
			name:   "Empty fields",
			input:  "a,,c",
			expect: []string{"a", "", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCSVLine(tt.input)
			if len(result) != len(tt.expect) {
				t.Errorf("parseCSVLine(%q) = %v, want %v", tt.input, result, tt.expect)
				return
			}
			for i, v := range result {
				if v != tt.expect[i] {
					t.Errorf("parseCSVLine(%q)[%d] = %q, want %q", tt.input, i, v, tt.expect[i])
				}
			}
		})
	}
}

// ===== PlantUML Encoding Tests =====

func TestEncodePlantUML(t *testing.T) {
	// Test that encoding produces a non-empty result
	input := "@startuml\nAlice -> Bob: Hello\n@enduml"
	result := encodePlantUML(input)

	if result == "" {
		t.Error("encodePlantUML should not return empty string")
	}

	// Result should only contain valid PlantUML alphabet characters
	validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	for _, c := range result {
		if !strings.ContainsRune(validChars, c) {
			t.Errorf("encodePlantUML result contains invalid character: %c", c)
		}
	}
}

func TestEncodePlantUMLConsistency(t *testing.T) {
	input := "@startuml\nA -> B\n@enduml"

	// Same input should produce same output
	result1 := encodePlantUML(input)
	result2 := encodePlantUML(input)

	if result1 != result2 {
		t.Error("encodePlantUML should be deterministic")
	}
}

// ===== Helper Function Tests =====

func TestSlugify(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"Hello World", "hello-world"},
		{"Test  Multiple   Spaces", "test-multiple-spaces"},
		{"Special!@#Characters", "special-characters"},
		{"Already-slugified", "already-slugified"},
		{"UPPERCASE", "uppercase"},
		{"émojis 🎉 here", "mojis-here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expect {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestReplaceEmojis(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{":smile:", "😊"},
		{":heart:", "❤️"},
		{":+1:", "👍"},
		{":unknown:", ":unknown:"},
		{"Hello :smile: World", "Hello 😊 World"},
		{"Multiple :heart: :star:", "Multiple ❤️ ⭐"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := replaceEmojis(tt.input)
			if result != tt.expect {
				t.Errorf("replaceEmojis(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

// ===== Text Rendering Tests =====

func TestRenderText(t *testing.T) {
	input := "Plain text content\nWith multiple lines"
	result := renderText(input)

	if !strings.Contains(result, `class="text"`) {
		t.Error("Should have text class")
	}
	if !strings.Contains(result, "search-toolbar") {
		t.Error("Should have search toolbar")
	}
	if !strings.Contains(result, "Plain text content") {
		t.Error("Should contain text content")
	}
}

// ===== Integration Tests =====

func TestRenderFile(t *testing.T) {
	// Test with non-existent file
	content, class := renderFile("/nonexistent/file.md")
	if !strings.Contains(content, "File not found") {
		t.Error("Should return file not found error")
	}
	if class != "" {
		t.Error("Class should be empty for error")
	}
}

// ===== Benchmark Tests =====

func BenchmarkRenderMarkdown(b *testing.B) {
	input := `# Header

This is a paragraph with **bold** and *italic* text.

## Code

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

## List

- Item 1
- Item 2
- Item 3

| Col1 | Col2 |
|------|------|
| A    | B    |
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderMarkdown(input, "")
	}
}

func BenchmarkRenderJSON(b *testing.B) {
	input := `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}], "count": 2}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderJSON(input)
	}
}

func BenchmarkEncodePlantUML(b *testing.B) {
	input := "@startuml\nAlice -> Bob: Hello\nBob --> Alice: Hi!\n@enduml"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodePlantUML(input)
	}
}
