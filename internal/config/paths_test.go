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
	os.Unsetenv(tenConfigEnvVar)
	t.Setenv(smartcatConfigEnvVar, "/tmp/smartcat")
	defer os.Unsetenv(smartcatConfigEnvVar)
	if got := ConfigDir(); got != "/tmp/smartcat" {
		t.Fatalf("want /tmp/smartcat, got %s", got)
	}
}

func TestConfigDirDefault(t *testing.T) {
	os.Unsetenv(tenConfigEnvVar)
	os.Unsetenv(smartcatConfigEnvVar)
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, defaultConfigDir)
	if got := ConfigDir(); got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
