package config

import (
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type (
	// Slack bot config details
	SlackBotConfig struct {
		BotToken  string `mapstructure:"slack_bot_token"`
		AppToken  string `mapstructure:"slack_app_token"`
		ChannelID string `mapstructure:"slack_channel_id"`
	}

	// Config defines all the app configurations
	Config struct {
		Slack SlackBotConfig `mapstructure:"slack"`
	}
)

// ReadConfigFromFile to read config details using viper
func ReadConfigFromFile() (*Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./config/")
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		fmt.Errorf("error while reading config.toml: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		fmt.Errorf("error unmarshaling config.toml to application config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Errorf("error occurred in config validation: %v", err)
	}

	return &cfg, nil
}

// Validate config struct
func (c *Config) Validate(e ...string) error {
	v := validator.New()
	if len(e) == 0 {
		return v.Struct(c)
	}
	return v.StructExcept(c, e...)
}
