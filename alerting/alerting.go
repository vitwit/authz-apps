package alerting

// SlackAlerter is an interface definition for slack related actions
// like Send slack alert
type SlackAlerter interface {
	Send(msgText, botToken, alertChannelID string) error
}

type slackAlert struct{}

// NewSlackAlerter returns a new instance for slackAlert
func NewSlackAlerter() *slackAlert {
	return &slackAlert{}
}
