package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirCustomEnv(t *testing.T) {
	t.Setenv(tenConfigEnvVar, "/tmp/custom_ten")
	t.Setenv(smartcatConfigEnvVar, "/tmp/smartcat")
	if got := ConfigDir(); got != "/tmp/custom_ten" {
		t.Fatalf("want /tmp/custom_ten, got %s", got)
	}
}

func TestConfigDirSmartcatFallback(t *testing.T) {
	os.Unsetenv(tenConfigEnvVar)          //nolint:errcheck
	t.Setenv(smartcatConfigEnvVar, "/tmp/smartcat")
	defer os.Unsetenv(smartcatConfigEnvVar) //nolint:errcheck
	if got := ConfigDir(); got != "/tmp/smartcat" {
		t.Fatalf("want /tmp/smartcat, got %s", got)
	}
}

func TestConfigDirDefault(t *testing.T) {
	os.Unsetenv(tenConfigEnvVar)      //nolint:errcheck
	os.Unsetenv(smartcatConfigEnvVar) //nolint:errcheck
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, defaultConfigDir)
	if got := ConfigDir(); got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestPromptsPath(t *testing.T) {
	t.Setenv(tenConfigEnvVar, "/tmp/x")
	want := "/tmp/x/prompts.toml"
	if got := PromptsPath(); got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestConversationPath(t *testing.T) {
	t.Setenv(tenConfigEnvVar, "/tmp/x")
	want := "/tmp/x/conversation.toml"
	if got := ConversationPath(); got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
