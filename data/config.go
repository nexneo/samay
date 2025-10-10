package data

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	configFileName = "config.json"
	defaultDBName  = "Samay.db"
)

type config struct {
	DatabasePath string `json:"database_path"`
}

// ResolveDatabasePath returns the persisted database location or prompts the user
// to choose one when Samay runs for the first time.
func ResolveDatabasePath() (string, error) {
	configDir, err := configDirectory()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0o775); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	cfgPath := filepath.Join(configDir, configFileName)
	if existing, err := readConfig(cfgPath); err == nil && existing.DatabasePath != "" {
		path, err := expandPath(existing.DatabasePath)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	defPath, err := defaultDatabasePath()
	if err != nil {
		return "", err
	}

	chosen, err := promptForPath(defPath)
	if err != nil {
		return "", err
	}

	if err := ensureParentDirectory(chosen); err != nil {
		return "", err
	}

	if err := writeConfig(cfgPath, config{DatabasePath: chosen}); err != nil {
		return "", err
	}

	return chosen, nil
}

func configDirectory() (string, error) {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "samay"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".samay"), nil
}

func defaultDatabasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	documents := filepath.Join(home, "Documents")
	return filepath.Join(documents, defaultDBName), nil
}

func readConfig(path string) (cfg config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return config{}, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close config file: %w", closeErr))
		}
	}()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return config{}, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func writeConfig(path string, cfg config) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close config file: %w", closeErr))
		}
	}()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func promptForPath(defaultPath string) (string, error) {
	fmt.Printf("Where should Samay store its database? [%s]: ", defaultPath)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read input: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		input = defaultPath
	}

	path, err := expandPath(input)
	if err != nil {
		return "", err
	}
	return path, nil
}

func expandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	} else if strings.HasPrefix(path, "~") {
		return "", fmt.Errorf("user-relative paths are not supported: %s", path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	return abs, nil
}

func ensureParentDirectory(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o775); err != nil {
		return fmt.Errorf("create parent directory %q: %w", dir, err)
	}
	return nil
}
