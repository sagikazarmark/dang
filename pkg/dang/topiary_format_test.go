package dang

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

func TestTopiaryFormatterParity(t *testing.T) {
	topiary, err := exec.LookPath("topiary")
	if err != nil {
		t.Skip("topiary is not installed")
	}
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("cc is not installed; cannot build Dang tree-sitter grammar for Topiary")
	}

	root := findRepoRoot(t)
	grammar := buildTopiaryGrammar(t, root)
	config := writeTopiaryConfig(t, grammar)
	query := filepath.Join(root, ".topiary", "queries", "dang.scm")

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "value assignment and call args",
			input: `pub x=foo(1,2)`,
		},
		{
			name:  "typed pub elision",
			input: `pub x: Int!=1`,
		},
		{
			name:  "private typed field",
			input: `let secret:String!="x"`,
		},
		{
			name:  "inline list",
			input: `pub x=[1,2,3]`,
		},
		{
			name: "multiline list",
			input: `pub x = [
  1
  2
]`,
		},
		{
			name:  "single line block arg",
			input: `pub x = items.map { x=>x+1 }`,
		},
		{
			name: "multiline chain",
			input: `pub x = foo
  .bar
  .baz`,
		},
		{
			name:  "single line enum",
			input: `enum Color{RED GREEN}`,
		},
		{
			name:  "union spacing",
			input: `union Shape=Circle|Square`,
		},
		{
			name:  "object declaration body",
			input: `type T{pub name:String!}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, err := FormatFile([]byte(tt.input))
			if err != nil {
				t.Fatalf("go formatter failed: %v", err)
			}

			got := runTopiaryFormat(t, topiary, config, query, tt.input)
			if got != expected {
				t.Fatalf("Topiary output differs from Go formatter\ninput:\n%s\nexpected:\n%s\ngot:\n%s", tt.input, expected, got)
			}
		})
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root")
		}
		dir = parent
	}
}

func buildTopiaryGrammar(t *testing.T, root string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "tree-sitter-dang.so")
	cmd := exec.Command("bash", filepath.Join(root, "hack", "build-topiary-grammar.sh"), out)
	cmd.Dir = root
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build Topiary grammar: %v\n%s", err, outBytes)
	}
	return out
}

func writeTopiaryConfig(t *testing.T, grammar string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "languages.ncl")
	config := fmt.Sprintf(`{
  languages.dang = {
    extensions = ["dang"],
    indent | force = "  ",
    grammar.source.path = %s,
  },
}
`, strconv.Quote(grammar))
	if err := os.WriteFile(path, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runTopiaryFormat(t *testing.T, topiary, config, query, input string) string {
	t.Helper()
	cmd := exec.Command(topiary, "format", "--language", "dang", "--configuration", config, "--query", query)
	cmd.Stdin = bytes.NewBufferString(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("topiary format failed: %v\n%s", err, out)
	}
	return string(out)
}
