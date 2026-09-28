package apns

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/Laboratory-404/mattermost-push-server/pkg/config"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// Provider implements Apple Push Notification service (APNs) HTTP/2 provider
type Provider struct {
	cfg        config.APNSConfig
	client     *apns2.Client
	mu         sync.RWMutex
}

// NewProvider initializes APNs HTTP/2 client
func NewProvider(cfg config.APNSConfig) (*Provider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	p := &Provider{
		cfg: cfg,
	}

	var client *apns2.Client

	// Option 1: JWT Auth Token (.p8)
	if cfg.AuthKeyFile != "" {
		cleanPath := filepath.Clean(cfg.AuthKeyFile)
		authKey, err := token.AuthKeyFromFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load APNs auth key file at %s: %w", cfg.AuthKeyFile, err)
		}

		tok := &token.Token{
			AuthKey: authKey,
			KeyID:   cfg.KeyID,
			TeamID:  cfg.TeamID,
		}

		if cfg.Production {
			client = apns2.NewTokenClient(tok).Production()
		} else {
			client = apns2.NewTokenClient(tok).Development()
		}

		log.Printf("[APNS] Configured APNs client using JWT Token (.p8 key_id: %s, team_id: %s, prod: %v)", cfg.KeyID, cfg.TeamID, cfg.Production)
	} else if cfg.CertFile != "" { // Option 2: PKCS12 Certificate (.p12)
		cleanPath := filepath.Clean(cfg.CertFile)
		cert, err := certificate.FromP12File(cleanPath, cfg.CertPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load APNs p12 cert at %s: %w", cfg.CertFile, err)
		}

		if cfg.Production {
			client = apns2.NewClient(cert).Production()
		} else {
			client = apns2.NewClient(cert).Development()
		}

		log.Printf("[APNS] Configured APNs client using PKCS12 (.p12 cert, prod: %v)", cfg.Production)
	} else {
		log.Printf("[APNS] APNs enabled but no AuthKeyFile (.p8) or CertFile (.p12) specified")
	}

	p.client = client
	return p, nil
}

func (p *Provider) Name() string {
	return "apns"
}

// Send dispatches notification to Apple APNs server
func (p *Provider) Send(ctx context.Context, msg *push.PushNotification) error {
	if p.client == nil {
		return fmt.Errorf("APNs client not configured")
	}

	if msg.DeviceID == "" {
		return fmt.Errorf("empty device_id")
	}

	title := msg.SenderName
	if title == "" {
		title = msg.ChannelName
	}

	// Build APNs Payload
	apnsPayload := payload.NewPayload().
		AlertTitle(title).
		AlertBody(msg.Message).
		Badge(msg.Badge).
		Sound(msg.Sound)

	if msg.Sound == "" {
		apnsPayload.Sound("default")
	}

	if msg.Category != "" {
		apnsPayload.Category(msg.Category)
	}

	// Attach custom data attributes required by Mattermost iOS App
	apnsPayload.Custom("channel_id", msg.ChannelID)
	apnsPayload.Custom("channel_name", msg.ChannelName)
	apnsPayload.Custom("team_id", msg.TeamID)
	apnsPayload.Custom("post_id", msg.PostID)
	apnsPayload.Custom("root_id", msg.RootID)
	apnsPayload.Custom("sender_id", msg.SenderID)
	apnsPayload.Custom("sender_name", msg.SenderName)
	apnsPayload.Custom("type", msg.Type)
	apnsPayload.Custom("version", msg.Platform)
	apnsPayload.Custom("override_username", msg.OverrideUsername)
	apnsPayload.Custom("override_icon_url", msg.OverrideIconURL)
	apnsPayload.Custom("from_webhook", msg.FromWebhook)
	apnsPayload.Custom("sub_type", msg.SubType)

	topic := p.cfg.Topic
	if topic == "" {
		// Default fallback bundle identifier for Mattermost React Native mobile app
		topic = "com.mattermost.rn"
	}

	notification := &apns2.Notification{
		DeviceToken: msg.DeviceID,
		Topic:       topic,
		Payload:     apnsPayload,
		Priority:    apns2.PriorityHigh,
	}

	res, err := p.client.PushWithContext(ctx, notification)
	if err != nil {
		return fmt.Errorf("APNs HTTP/2 request failed: %w", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("APNs error status %d: %s (Reason: %s)", res.StatusCode, res.ApnsID, res.Reason)
	}

	log.Printf("[APNS] Successfully sent APNs push notification to token %s...", truncate(msg.DeviceID, 12))
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
