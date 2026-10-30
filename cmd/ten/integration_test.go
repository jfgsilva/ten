package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is set once by TestMain.
var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary into a temp dir before running any tests.
	tmp, err := os.MkdirTemp("", "ten-integration-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmp) //nolint:errcheck

	binaryPath = filepath.Join(tmp, "ten")
	out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput()
	if err != nil {
		panic("failed to build ten: " + string(out))
	}

	os.Exit(m.Run()) //nolint:gocritic
}

// runTen executes the ten binary with TEN_TEST=1 (mock API mode),
// a temp config dir, the given stdin, and the provided args.
func runTen(t *testing.T, stdinData string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	configDir := t.TempDir()
	// Write minimal config files so EnsureConfigFiles doesn't bail.
	writeTestConfigs(t, configDir)

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = strings.NewReader(stdinData)
	cmd.Env = append(os.Environ(),
		"TEN_TEST=1",
		"TEN_CONFIG_PATH="+configDir,
	)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// writeTestConfigs creates minimal prompts.toml and .api_configs.toml.
func writeTestConfigs(t *testing.T, dir string) {
	t.Helper()

	prompts := `[default]
api = "ollama"
char_limit = 50000

[[default.messages]]
role = "system"
content = "You are a smart cat."
`
	if err := os.WriteFile(filepath.Join(dir, "prompts.toml"), []byte(prompts), 0o644); err != nil {
		t.Fatal(err)
	}

	apiCfg := `[ollama]
url = "http://localhost:11434/api/chat"
default_model = "phi3"
timeout_seconds = 180
`
	if err := os.WriteFile(filepath.Join(dir, ".api_configs.toml"), []byte(apiCfg), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPipeInputWithPromptArg mirrors: git diff | ten "summarize changes in 1 sentence"
func TestPipeInputWithPromptArg(t *testing.T) {
	diffOutput := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+
+// added a comment
 func main() {}`

	stdout, _, err := runTen(t, diffOutput, "summarize changes in 1 sentence")
	if err != nil {
		t.Fatalf("ten exited with error: %v", err)
	}

	// TEN_TEST=1 echoes stdin back; verify the piped diff appears in output.
	if !strings.Contains(stdout, "added a comment") {
		t.Errorf("expected piped git diff in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Hello, World!") {
		t.Errorf("expected test mode header in output, got:\n%s", stdout)
	}
}

// TestPipeInputNoArgs mirrors: git diff | ten
func TestPipeInputNoArgs(t *testing.T) {
	diffOutput := "diff --git a/foo.go b/foo.go\n+added line"

	stdout, _, err := runTen(t, diffOutput)
	if err != nil {
		t.Fatalf("ten exited with error: %v", err)
	}
	if !strings.Contains(stdout, "added line") {
		t.Errorf("expected piped input in output, got:\n%s", stdout)
	}
}

// TestPipeEmptyStdinWithArg mirrors: ten "just a question"
// TEN_TEST=1 only echoes stdin, so we pass the question as stdin to simulate
// the real flow where an empty-pipe arg becomes the input.
func TestPipeEmptyStdinWithArg(t *testing.T) {
	stdout, _, err := runTen(t, "just a question", "summarize this")
	if err != nil {
		t.Fatalf("ten exited with error: %v", err)
	}
	if !strings.Contains(stdout, "just a question") {
		t.Errorf("expected stdin content in output, got:\n%s", stdout)
	}
}

// TestRepeatInput mirrors: git diff | ten -r "summarize changes"
func TestRepeatInput(t *testing.T) {
	// TEN_TEST=1 exits before -r flag is processed, so test that the flag
	// is accepted without error (flag parsing happens before TEN_TEST check
	// would be needed — in test mode we exit early, so just ensure no crash).
	_, _, err := runTen(t, "some input", "-r", "summarize changes")
	if err != nil {
		t.Fatalf("ten -r exited with error: %v", err)
	}
}

// TestTemplateRefThenInput mirrors: ten default "my input"
// Uses the "default" template (present in our test config) with piped input.
func TestTemplateRefThenInput(t *testing.T) {
	stdout, _, err := runTen(t, "my piped input", "default")
	if err != nil {
		t.Fatalf("ten exited with error: %v", err)
	}
	if !strings.Contains(stdout, "my piped input") {
		t.Errorf("expected piped input in output, got:\n%s", stdout)
	}
}

// TestLargePipeInput verifies multi-line piped input works end-to-end.
func TestLargePipeInput(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("diff line content here\n")
	}
	bigDiff := sb.String()

	stdout, _, err := runTen(t, bigDiff, "summarize changes")
	if err != nil {
		t.Fatalf("ten exited with error: %v", err)
	}
	if !strings.Contains(stdout, "diff line content here") {
		t.Errorf("expected piped content in output, got:\n%s", stdout[:min(len(stdout), 200)])
	}
}
