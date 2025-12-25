// Copyright (c) 2025 JoeGlenn1213
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

const (
	// DefaultPort is the default HTTP server port
	DefaultPort = 9418
	// DefaultBindAddress is the default bind address
	DefaultBindAddress = "127.0.0.1"
	// ConfigFileName is the name of the config file
	ConfigFileName = "config"
	// ConfigFileType is the type of the config file
	ConfigFileType = "yaml"
)

var (
	once     sync.Once
	instance *Config
)

// Config holds the application configuration
type Config struct {
	Port        int    `mapstructure:"port"`
	BindAddress string `mapstructure:"bind_address"`
	ReposDir    string `mapstructure:"repos_dir"`
	ReadOnly    bool   `mapstructure:"read_only"`
	MDNSEnabled bool   `mapstructure:"mdns_enabled"`
	DataDir     string `mapstructure:"data_dir"`
	// Authentication (optional)
	AuthEnabled      bool   `mapstructure:"auth_enabled"`
	AuthUser         string `mapstructure:"auth_user"`
	AuthPasswordHash string `mapstructure:"auth_password_hash"`
}

// GetLGHDir returns the LGH data directory path
func GetLGHDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".localgithub"
	}
	return filepath.Join(home, ".localgithub")
}

// GetReposDir returns the repos directory path
func GetReposDir() string {
	return filepath.Join(GetLGHDir(), "repos")
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	return filepath.Join(GetLGHDir(), ConfigFileName+"."+ConfigFileType)
}

// GetMappingsPath returns the mappings file path
func GetMappingsPath() string {
	return filepath.Join(GetLGHDir(), "mappings.yaml")
}

// GetPIDPath returns the PID file path
func GetPIDPath() string {
	return filepath.Join(GetLGHDir(), "lgh.pid")
}

// Load loads the configuration from file
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		instance = &Config{
			Port:        DefaultPort,
			BindAddress: DefaultBindAddress,
			ReposDir:    GetReposDir(),
			ReadOnly:    false,
			MDNSEnabled: false,
			DataDir:     GetLGHDir(),
		}

		viper.SetConfigName(ConfigFileName)
		viper.SetConfigType(ConfigFileType)
		viper.AddConfigPath(GetLGHDir())

		// Set defaults
		viper.SetDefault("port", DefaultPort)
		viper.SetDefault("bind_address", DefaultBindAddress)
		viper.SetDefault("repos_dir", GetReposDir())
		viper.SetDefault("read_only", false)
		viper.SetDefault("mdns_enabled", false)
		viper.SetDefault("data_dir", GetLGHDir())

		if readErr := viper.ReadInConfig(); readErr != nil {
			if _, ok := readErr.(viper.ConfigFileNotFoundError); !ok {
				err = fmt.Errorf("error reading config file: %w", readErr)
				return
			}
			// Config file not found, use defaults
		}

		if unmarshalErr := viper.Unmarshal(instance); unmarshalErr != nil {
			err = fmt.Errorf("error unmarshaling config: %w", unmarshalErr)
			return
		}
	})

	return instance, err
}

// Get returns the cached configuration instance
func Get() *Config {
	if instance == nil {
		cfg, _ := Load()
		return cfg
	}
	return instance
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	viper.Set("port", cfg.Port)
	viper.Set("bind_address", cfg.BindAddress)
	viper.Set("repos_dir", cfg.ReposDir)
	viper.Set("read_only", cfg.ReadOnly)
	viper.Set("mdns_enabled", cfg.MDNSEnabled)
	viper.Set("data_dir", cfg.DataDir)

	configPath := GetConfigPath()
	return viper.WriteConfigAs(configPath)
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig() error {
	cfg := &Config{
		Port:        DefaultPort,
		BindAddress: DefaultBindAddress,
		ReposDir:    GetReposDir(),
		ReadOnly:    false,
		MDNSEnabled: false,
		DataDir:     GetLGHDir(),
	}
	return Save(cfg)
}

// Reset clears the singleton for testing
func Reset() {
	once = sync.Once{}
	instance = nil
}
