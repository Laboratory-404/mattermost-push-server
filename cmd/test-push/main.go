package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8066", "Push Proxy Server URL")
	platform := flag.String("platform", "android_rn", "Target platform (android_rn, apple_rn, web, desktop)")
	deviceToken := flag.String("token", "test_device_token_sample", "Target FCM / APNs device token")
	message := flag.String("message", "Test push notification from Mattermost Push Proxy!", "Notification text")
	sender := flag.String("sender", "System Administrator", "Sender display name")
	flag.Parse()

	fmt.Println("==================================================")
	fmt.Println(" Mattermost Push Proxy Test Utility")
	fmt.Println("==================================================")

	payload := push.PushNotification{
		Platform:     *platform,
		DeviceID:     *deviceToken,
		Type:         "message",
		ChannelID:    "test_channel_id",
		ChannelName:  "General",
		TeamID:       "test_team_id",
		SenderID:     "sysadmin_user_id",
		SenderName:   *sender,
		Message:      *message,
		Badge:        1,
		PostID:       "test_post_id",
		Sound:        "default",
		ReceivedAt:   time.Now(),
	}

	bodyBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	targetURL := *serverURL + "/api/v1/send_push"
	fmt.Printf("Sending test notification payload to: %s\n", targetURL)
	fmt.Printf("Payload:\n%s\n\n", string(bodyBytes))

	resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Server Response: %s\n", string(respBody))

	if resp.StatusCode == http.StatusOK {
		var res push.PushResponse
		if err := json.Unmarshal(respBody, &res); err == nil && res.Status == "OK" {
			fmt.Println("\n✅ Push Notification Test Succeeded!")
			os.Exit(0)
		}
	}

	fmt.Println("\n❌ Push Notification Test Failed!")
	os.Exit(1)
}
