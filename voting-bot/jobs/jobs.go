package jobs

import (
	"log"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/vitwit/authz-apps/voting-bot/types"
)

// Cron wraps all required parameters to create cron jobs
type Cron struct {
	ctx    types.Context
	logger *zerolog.Logger
}

// NewCron sets necessary config and clients to begin cron jobs
func NewCron(ctx types.Context) *Cron {
	return &Cron{
		ctx:    ctx,
		logger: ctx.Logger(),
	}
}

// Start starts to create cron jobs which sends alerts on proposal alerts which have not been voted
func (c *Cron) Start() error {
	c.logger.Info().Msg("Starting cron jobs...")

	cron := cron.New()

	// Everday at 8AM and 8PM
	_, err := cron.AddFunc("0 8,20 * * *", func() {
		GetProposals(c.ctx)
		GetLowBalAccs(c.ctx)
	})
	if err != nil {
		log.Println("Error while adding Proposals and Low balance accounts alerting cron jobs:", err)
		return err
	}
	_, err = cron.AddFunc("@every 24h", func() {
		SyncAuthzStatus(c.ctx)
	})
	if err != nil {
		log.Println("Error while adding Key Authorization syncing cron job:", err)
		return err
	}

	_, err = cron.AddFunc("@every 2h", func() {
		log.Printf("Running withdraw commission CRON job....")
		if err := Withdraw(c.ctx); err != nil {
			log.Println("Error while adding Key Authorization syncing cron job:", err)
		}
	})
	if err != nil {
		log.Println("Error while adding withdraw rewards and commission cron job:", err)
		return err
	}

	go cron.Start()

	return nil
}
