package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ribice/goch"
	"gopkg.in/yaml.v2"
)

// Config represents application configuration
type Config struct {
	Server    *Server               `yaml:"server,omitempty"`
	Redis     *Redis                `yaml:"redis,omitempty"`
	NATS      *NATS                 `yaml:"nats,omitempty"`
	Admin     *AdminAccount         `yaml:"-"`
	Limits    map[goch.Limit][2]int `yaml:"limits,omitempty"`
	LimitErrs map[goch.Limit]error  `yaml:"-"`
}

// Server holds data necessery for server configuration
type Server struct {
	Port int `yaml:"port"`
}

// Redis holds credentials for Redis
type Redis struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Password string `yaml:"-"`
}

// NATS holds credentials for NATS-Streaming server
type NATS struct {
	ClusterID string `yaml:"cluster_id"`
	ClientID  string `yaml:"client_id"`
	URL       string `yaml:"url"`
}

// AdminAccount represents an account needed for creating new channels
type AdminAccount struct {
	Username string
	Password string
}

// Load loads config from file and env variables
func Load(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file, %s", err)
	}
	var cfg = new(Config)
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct, %v", err)
	}

	if len(cfg.Limits) != int(goch.ChanSecretLimit) {
		return nil, fmt.Errorf("not all limits were loaded. needed %v, actual %v", goch.ChanSecretLimit, len(cfg.Limits))
	}

	if cfg.Redis != nil {
		cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}

	user, err := getEnv("ADMIN_USERNAME")
	if err != nil {
		return nil, err
	}

	pass, err := getEnv("ADMIN_PASSWORD")
	if err != nil {
		return nil, err
	}

	cfg.Admin = &AdminAccount{Username: user, Password: pass}
	cfg.LimitErrs = map[goch.Limit]error{
		goch.DisplayNameLimit: fmt.Errorf("displayName must be between %v and %v characters long", cfg.Limits[1][0], cfg.Limits[1][1]),
		goch.UIDLimit:         fmt.Errorf("uid must be between %v and %v characters long", cfg.Limits[2][0], cfg.Limits[2][1]),
		goch.SecretLimit:      fmt.Errorf("secret must be between %v and %v characters long", cfg.Limits[3][0], cfg.Limits[3][1]),
		goch.ChanLimit:        fmt.Errorf("channel must be between %v and %v characters long", cfg.Limits[4][0], cfg.Limits[4][1]),
		goch.ChanSecretLimit:  fmt.Errorf("channelSecret must be between %v and %v characters long", cfg.Limits[5][0], cfg.Limits[5][1]),
	}
	return cfg, nil
}

// ExceedsAny checks whether any limit is exceeded
func (c *Config) ExceedsAny(m map[string]goch.Limit) error {
	for k, v := range m {
		if exceedsLim(k, c.Limits[v]) {
			return c.LimitErrs[v]
		}
	}
	return nil

}

// Exceeds checks whether a string exceeds chat limitation
func (c *Config) Exceeds(str string, lim goch.Limit) error {
	if exceedsLim(str, c.Limits[lim]) {
		return c.LimitErrs[lim]
	}
	return nil
}

func exceedsLim(s string, lims [2]int) bool {
	return len(s) < lims[0] || len(s) > lims[1]
}

func getEnv(key string) (string, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return "", fmt.Errorf("env variable %s required but not found", key)
	}
	return v, nil
}
