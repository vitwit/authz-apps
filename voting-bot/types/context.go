package types

import (
	"context"

	"github.com/shomali11/slacker"
	"github.com/vitwit/authz-apps/voting-bot/config"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rs/zerolog"

	registry "github.com/strangelove-ventures/lens/client/chain_registry"
)

type Context struct {
	baseCtx context.Context
	logger  zerolog.Logger

	database *database.Sqlitedb
	cfg      *config.Config
	slacker  *slacker.Slacker

	chainRegistry registry.ChainRegistry
}

// create a new context
func NewContext(
	logger zerolog.Logger, database *database.Sqlitedb, cfg *config.Config, slacker *slacker.Slacker,
) Context {
	return Context{
		baseCtx:       context.Background(),
		logger:        logger,
		database:      database,
		cfg:           cfg,
		slacker:       slacker,
		chainRegistry: registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore())),
	}
}

// WithContext returns a Context with an updated context.Context.
func (c Context) WithContext(ctx context.Context) Context {
	c.baseCtx = ctx
	return c
}

// WithLogger returns a Context with an updated logger.
func (c Context) WithLogger(l zerolog.Logger) Context {
	c.logger = l
	return c
}

// WithConfig returns a Context with an updated config.
func (c Context) WithConfig(cfg *config.Config) Context {
	c.cfg = cfg
	return c
}

// WithSlackClient returns a Context with an updated slack client.
func (c Context) WithSlackClient(client *slacker.Slacker) Context {
	c.slacker = client
	return c
}

func (c Context) Context() context.Context {
	return c.baseCtx
}

func (c Context) Slacker() *slacker.Slacker {
	return c.slacker
}

func (c Context) Database() *database.Sqlitedb {
	return c.database
}

func (c Context) Config() *config.Config {
	return c.cfg
}

func (c Context) Logger() *zerolog.Logger {
	return &c.logger
}

func (c Context) ChainRegistry() registry.ChainRegistry {
	return c.chainRegistry
}
