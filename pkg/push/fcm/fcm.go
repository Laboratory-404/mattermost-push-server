package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Laboratory-404/mattermost-push-server/pkg/config"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	fcmScope = "https://www.googleapis.com/auth/firebase.messaging"
)

// Provider implements Firebase Cloud Messaging HTTP v1 push provider
type Provider struct {
	cfg        config.FCMConfig
	httpClient *http.http.Client
	tokenSource oauth2.TokenSource
	projectID  string
	mu         sync.RWMutex
}

// ServiceAccountFile structure to extract project_id if needed
type serviceAccountJSON struct {
	ProjectID string `json:"project_id"`
}

// NewProvider initializes FCM provider using HTTP v1 API
func NewProvider(cfg config.FCMConfig) (*Provider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	p := &Provider{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		projectID:  cfg.ProjectID,
	}

	if cfg.CredentialsFile != "" {
		cleanPath := filepath.Clean(cfg.CredentialsFile)
		jsonBytes, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read FCM credentials file %s: %w", cfg.CredentialsFile, err)
		}

		// Extract project_id if missing
		if p.projectID == "" {
			var sa serviceAccountJSON
			if err := json.Unmarshal(jsonBytes, &sa); err == nil && sa.ProjectID != "" {
				p.projectID = sa.ProjectID
			}
		}

		creds, err := google.CredentialsFromJSON(context.Background(), jsonBytes, fcmScope)
		if err != nil {
			return nil, fmt.Errorf("failed to parse FCM credentials JSON: %w", err)
		}

		p.tokenSource = creds.TokenSource
		log.Printf("[FCM] Configured FCM HTTP v1 for project: %s", p.projectID)
	} else if cfg.AndroidLegacyKey != "" {
		log.Printf("[FCM] Warning: Credentials file not provided, using Legacy FCM API key fallback")
	} else {
		log.Printf("[FCM] FCM enabled but no credentials file or legacy key specified")
	}

	return p, nil
}

func (p *Provider) Name() string {
	return "fcm"
}

// Send sends push notification to FCM
func (p *Provider) Send(ctx context.Context, msg *push.PushNotification) error {
	if msg.DeviceID == "" {
		return fmt.Errorf("empty device_id")
	}

	if p.tokenSource != nil && p.projectID != "" {
		return p.sendHTTPv1(ctx, msg)
	}

	if p.cfg.AndroidLegacyKey != "" {
		return p.sendLegacy(ctx, msg)
	}

	return fmt.Errorf("FCM credentials not initialized")
}

// sendHTTPv1 sends notification via FCM HTTP v1 API
func (p *Provider) sendHTTPv1(ctx context.Context, msg *push.PushNotification) error {
	token, err := p.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get FCM oauth2 token: %w", err)
	}

	title := msg.SenderName
	if title == "" {
		title = msg.ChannelName
	}
	if title == "" {
		title = "Mattermost"
	}

	// FCM HTTP v1 Payload Structure
	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": msg.DeviceID,
			"notification": map[string]string{
				"title": title,
				"body":  msg.Message,
			},
			"data": map[string]string{
				"channel_id":        msg.ChannelID,
				"channel_name":      msg.ChannelName,
				"team_id":           msg.TeamID,
				"post_id":           msg.PostID,
				"root_id":           msg.RootID,
				"sender_id":         msg.SenderID,
				"sender_name":       msg.SenderName,
				"type":              msg.Type,
				"badge":             strconv.Itoa(msg.Badge),
				"version":           msg.Platform,
				"override_username": msg.OverrideUsername,
				"override_icon_url": msg.OverrideIconURL,
				"from_webhook":      msg.FromWebhook,
				"sub_type":          msg.SubType,
			},
			"android": map[string]interface{}{
				"priority": "HIGH",
				"notification": map[string]interface{}{
					"sound":                  "default",
					"click_action":           "TOP_STORY_ACTIVITY",
					"notification_priority": "PRIORITY_HIGH",
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM HTTP v1 request: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", p.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create FCM request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("FCM HTTP v1 request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM API error status %d: %s", resp.StatusCode, string(respBytes))
	}

	log.Printf("[FCM] Successfully sent HTTP v1 push notification to device %s...", truncate(msg.DeviceID, 12))
	return nil
}

// sendLegacy fallback for legacy FCM endpoint if needed
func (p *Provider) sendLegacy(ctx context.Context, msg *push.PushNotification) error {
	payload := map[string]interface{}{
		"to": msg.DeviceID,
		"data": map[string]interface{}{
			"message":      msg.Message,
			"channel_id":   msg.ChannelID,
			"channel_name": msg.ChannelName,
			"team_id":      msg.TeamID,
			"post_id":      msg.PostID,
			"type":         msg.Type,
			"badge":        msg.Badge,
			"sender_id":    msg.SenderID,
			"sender_name":  msg.SenderName,
			"version":      msg.Platform,
		},
		"priority": "high",
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "key="+p.cfg.AndroidLegacyKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Legacy FCM error status %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
