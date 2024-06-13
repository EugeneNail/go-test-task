package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Symbols    []string `yaml:"symbols"`
	MaxWorkers int      `yaml:"max_workers"`
}

func Load() (Config, error) { // лучше путь передавать сюда параметром - это будет более расширяемое
	var config Config

	directory, err := getProjectDirectory() // можно конечно, но сложнаааа.
	if err != nil {
		return config, fmt.Errorf("config.Load: %w", err) // лучше errors.Wrap - так окнечно можно, но обычно я везде видел  явный Wrap
	}

	file, err := os.ReadFile(filepath.Join(directory, "config.yaml"))
	if err != nil {
		return config, fmt.Errorf("config.Load: %w", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, fmt.Errorf("config.Load: %w", err)
	}

	config.MaxWorkers = min(config.MaxWorkers, runtime.NumCPU())

	return config, nil
}

func getProjectDirectory() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getProjectDirectory: %w", err)
	}

	for {
		_, err := os.Stat(filepath.Join(directory, "go.mod"))
		if err == nil {
			return directory, nil
		}

		if directory == filepath.Dir(directory) {
			return "", errors.New("getProjectDirectory: could not find the go.mod file")
		}
		directory = filepath.Dir(directory)
	}
}
