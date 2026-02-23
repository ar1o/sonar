package config

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const dbFileName = "issues.db"

// Config holds resolved configuration for the sonar directory and database.
type Config struct {
	SonarDir string // resolved .sonar directory path
	DBPath    string // full path to issues.db
	EnvVarSet bool   // whether SONAR_PATH was used
}

// Resolve returns the current configuration by checking SONAR_PATH first,
// then falling back to $PWD/.sonar.
func Resolve() (*Config, error) {
	var sonarDir string
	var envVarSet bool

	if envPath := os.Getenv("SONAR_PATH"); envPath != "" {
		sonarDir = envPath
		envVarSet = true
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		sonarDir = filepath.Join(cwd, ".sonar")
	}

	return &Config{
		SonarDir: sonarDir,
		DBPath:    filepath.Join(sonarDir, dbFileName),
		EnvVarSet: envVarSet,
	}, nil
}

// Exists checks if the sonar directory and DB file both exist.
// It returns an error for non-existence failures (e.g. permission errors).
func (c *Config) Exists() (bool, error) {
	if _, err := os.Stat(c.SonarDir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if _, err := os.Stat(c.DBPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

var (
	defaultAuthor     string
	defaultAuthorOnce sync.Once
)

// DefaultAuthor returns the default author for comments and activity.
// It tries git config user.name first and falls back to the OS username.
// The result is cached for the lifetime of the process.
func DefaultAuthor() string {
	defaultAuthorOnce.Do(func() {
		defaultAuthor = resolveAuthor()
	})
	return defaultAuthor
}

func resolveAuthor() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "config", "user.name").Output()
	if err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}

	u, err := user.Current()
	if err == nil && u.Username != "" {
		return u.Username
	}

	return "unknown"
}
