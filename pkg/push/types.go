package push

import (
	"context"
	"time"
)

// Target platforms supported by Mattermost push requests
const (
	PlatformApple     = "apple"
	PlatformAppleRN   = "apple_rn"
	PlatformAndroid   = "android"
	PlatformAndroidRN = "android_rn"
	PlatformWeb       = "web"
	PlatformDesktop   = "desktop"
)

// Notification types
const (
	TypeMessage  = "message"
	TypeClear    = "clear"
	TypeTest     = "test"
	TypeIdLoaded = "id_loaded"
)

// PushNotification represents the standard payload received from a Mattermost server
type PushNotification struct {
	Platform         string    `json:"version"`            // "apple", "apple_rn", "android", "android_rn", "web", "desktop"
	DeviceID         string    `json:"device_id"`          // Target FCM / APNs / Web Push Token
	Category         string    `json:"category,omitempty"` // iOS category or notification category
	Type             string    `json:"type"`               // "message", "clear", "test"
	ChannelID        string    `json:"channel_id"`         // ID of channel
	ChannelName      string    `json:"channel_name"`       // Channel display name
	TeamID           string    `json:"team_id"`            // ID of team
	SenderID         string    `json:"sender_id"`          // User ID of sender
	SenderName       string    `json:"sender_name"`        // Display name of sender
	Message          string    `json:"message"`            // Notification message text
	Badge            int       `json:"badge"`              // Icon badge count
	PostID           string    `json:"post_id"`            // ID of post
	RootID           string    `json:"root_id"`            // ID of root post in thread
	OverrideUsername string    `json:"override_username"`  // Webhook override username
	OverrideIconURL  string    `json:"override_icon_url"`  // Webhook override icon URL
	FromWebhook      string    `json:"from_webhook"`       // "true" or "false"
	Sound            string    `json:"sound"`              // Sound identifier
	SubType          string    `json:"sub_type"`           // Notification sub-type
	IsIDLoaded       bool      `json:"is_id_loaded"`       // ID-only payload flag
	ReceivedAt       time.Time `json:"-"`                  // Server timestamp when received
}

// PushResponse represents the response returned to Mattermost server
type PushResponse struct {
	Status string `json:"status"`          // "OK" or "FAIL"
	Error  string `json:"error,omitempty"` // Error details if status == "FAIL"
}

// Provider defines the standard interface for push notification transport backends (FCM, APNs, Web)
type Provider interface {
	// Send dispatches the notification to the target device/token
	Send(ctx context.Context, msg *PushNotification) error
	// Name returns the provider identifier
	Name() string
}
