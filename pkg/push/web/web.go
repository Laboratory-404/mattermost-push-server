package web

import (
	"context"
	"log"

	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
)

// Provider handles Web Push and Desktop notification bridging
type Provider struct{}

// NewProvider returns a new Web Push provider instance
func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "web"
}

// Send logs/dispatches web push notification payloads
func (p *Provider) Send(ctx context.Context, msg *push.PushNotification) error {
	log.Printf("[WEB/DESKTOP] Received notification for platform '%s' (Device: %s): Sender=%s, Msg=%s",
		msg.Platform, msg.DeviceID, msg.SenderName, msg.Message)
	return nil
}
