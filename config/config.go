package config

import (
	"log"

	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type (
	// Slack bot config details
	SlackBotConfig struct {
		BotToken  string `mapstructure:"SLACK_BOT_TOKEN"`
		AppToken  string `mapstructure:"SLACK_APP_TOKEN"`
		ChannelID string `mapstructure:"CHANNEL_ID"`
	}

	// // VotingPeriodAlert defines about voting period alerts
	// VotingPeriodAlert struct {
	// 	EnableAlert   string `mapstructure:"enable_alert"`
	// 	VotingEndTime string `mapstructure:"voting_end_time"`
	// }

	// Config defines all the app configurations
	Config struct {
		EnableSlackAlerts string         `mapstructure:"enable_slack_alerts"`
		Slack             SlackBotConfig `mapstructure:"slack"`
		VotingPeriodAlert string         `mapstructure:"voting_period_alert"`
	}
)

// ReadConfigFromFile to read config details using viper
func ReadConfigFromFile() (*Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./config/")
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error while reading config.toml: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshaling config.toml to application config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("error occurred in config validation: %v", err)
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
