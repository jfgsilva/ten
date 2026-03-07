package config

import (
	"os"
	"path/filepath"
)

const (
	defaultConfigDir     = ".config/ten"
	tenConfigEnvVar      = "TEN_CONFIG_PATH"
	smartcatConfigEnvVar = "SMARTCAT_CONFIG_PATH" // backward compat alias

	apiConfigFile   = ".api_configs.toml"
	promptsFile     = "prompts.toml"
	conversationFile = "conversation.toml"
)

// ConfigDir returns the config directory path.
// Priority: TEN_CONFIG_PATH > SMARTCAT_CONFIG_PATH > $HOME/.config/ten
func ConfigDir() string {
	if v := os.Getenv(tenConfigEnvVar); v != "" {
		return v
	}
	if v := os.Getenv(smartcatConfigEnvVar); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot determine home directory: " + err.Error())
	}
	return filepath.Join(home, defaultConfigDir)
}

func APIConfigPath() string {
	return filepath.Join(ConfigDir(), apiConfigFile)
}

func PromptsPath() string {
	return filepath.Join(ConfigDir(), promptsFile)
}

func ConversationPath() string {
	return filepath.Join(ConfigDir(), conversationFile)
}
