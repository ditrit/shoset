package shoset

import (
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	baseDir string
	viper   *viper.Viper
	mu      sync.Mutex
}

// NewConfig : constructor
func NewConfig() *Config {
	homeDir := "."
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error().Msg("couldn't home dir folder: " + err.Error())
	}
	cfg := &Config{
		baseDir: homeDir + "/.shoset/",
		viper:   viper.New(),
	}
	if err := maybeMkdir(cfg.baseDir); err != nil {
		log.Error().Msg("couldn't create folder: " + err.Error())
	}
	return cfg
}

func (cfg *Config) GetBaseDir() string {
	return cfg.baseDir
}

// InitFolders creates following config folders if needed:
// - <base>/<name>/{cert, config}
func (cfg *Config) InitFolders(name string) (string, error) {
	if err := maybeMkdir(cfg.baseDir + name + "/"); err != nil {
		return "", err
	}
	if err := maybeMkdir(cfg.baseDir + name + "/config/"); err != nil {
		return "", err
	}
	if err := maybeMkdir(cfg.baseDir + name + "/cert/"); err != nil {
		return "", err
	}

	return cfg.baseDir, nil
}

// ReadConfig wraps around viper.ReadInConfig with extra init steps.
func (cfg *Config) ReadConfig(name string) error {
	cfg.viper.AddConfigPath(cfg.baseDir + name + "/config/")
	cfg.viper.SetConfigName(name)
	cfg.viper.SetConfigType("yaml")

	return cfg.viper.ReadInConfig()
}

// WriteConfig wraps around viper.WriteConfigAs.
func (cfg *Config) WriteConfig(name string) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return cfg.viper.WriteConfigAs(cfg.baseDir + name + "/config/config.yaml")
}

func (cfg *Config) Set(key string, value interface{}) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.viper.Set(key, value)
}

func (cfg *Config) GetSlice(key string) []string {
	return cfg.viper.GetStringSlice(key)
}
