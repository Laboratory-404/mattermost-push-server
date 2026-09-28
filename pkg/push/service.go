package push

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Service coordinates routing of push notifications to providers
type Service struct {
	fcmProvider  Provider
	apnsProvider Provider
	webProvider  Provider
}

// NewService creates a new Push Notification Service
func NewService(fcm Provider, apns Provider, web Provider) *Service {
	return &Service{
		fcmProvider:  fcm,
		apnsProvider: apns,
		webProvider:  web,
	}
}

// Dispatch routes the incoming push notification request to the correct provider
func (s *Service) Dispatch(ctx context.Context, msg *PushNotification) error {
	msg.ReceivedAt = time.Now()
	platform := strings.ToLower(strings.TrimSpace(msg.Platform))

	log.Printf("[PushService] Dispatching push for type='%s', platform='%s', channel='%s', device_id_prefix='%s'",
		msg.Type, msg.Platform, msg.ChannelID, truncateID(msg.DeviceID, 10))

	switch platform {
	case PlatformAndroid, PlatformAndroidRN:
		if s.fcmProvider == nil {
			return fmt.Errorf("FCM provider for Android notifications is not configured or disabled")
		}
		return s.fcmProvider.Send(ctx, msg)

	case PlatformApple, PlatformAppleRN:
		if s.apnsProvider == nil {
			return fmt.Errorf("APNs provider for iOS notifications is not configured or disabled")
		}
		return s.apnsProvider.Send(ctx, msg)

	case PlatformWeb, PlatformDesktop:
		if s.webProvider != nil {
			return s.webProvider.Send(ctx, msg)
		}
		log.Printf("[PushService] Web/Desktop notification received: %s", msg.Message)
		return nil

	default:
		// Attempt fallback detection based on token format or default to FCM/APNs if only one is enabled
		if s.fcmProvider != nil && (strings.HasPrefix(msg.DeviceID, "fcm:") || strings.Contains(msg.DeviceID, ":APA91b")) {
			return s.fcmProvider.Send(ctx, msg)
		}
		if s.apnsProvider != nil && len(msg.DeviceID) == 64 {
			return s.apnsProvider.Send(ctx, msg)
		}

		return fmt.Errorf("unsupported or unrecognized target platform '%s'", msg.Platform)
	}
}

func truncateID(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
