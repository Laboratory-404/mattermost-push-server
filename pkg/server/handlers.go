package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Laboratory-404/mattermost-push-server/pkg/metrics"
	"github.com/Laboratory-404/mattermost-push-server/pkg/push"
)

// HandlerContainer holds service references for HTTP routing
type HandlerContainer struct {
	pushService *push.Service
}

// NewHandlerContainer initializes container with push service
func NewHandlerContainer(ps *push.Service) *HandlerContainer {
	return &HandlerContainer{pushService: ps}
}

// HandleSendPush handles POST /api/v1/send_push and /api/v2/send_push requests from Mattermost server
func (h *HandlerContainer) HandleSendPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		metrics.RecordRequest("unknown", false)
		writeJSONResponse(w, http.StatusBadRequest, push.PushResponse{
			Status: "FAIL",
			Error:  "Failed to read request body: " + err.Error(),
		})
		return
	}
	defer r.Body.Close()

	var notification push.PushNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		metrics.RecordRequest("unknown", false)
		writeJSONResponse(w, http.StatusBadRequest, push.PushResponse{
			Status: "FAIL",
			Error:  "Invalid JSON payload: " + err.Error(),
		})
		return
	}

	// Dispatch notification via Push Service
	if err := h.pushService.Dispatch(r.Context(), &notification); err != nil {
		log.Printf("[Handler] Send push failed for device %s (platform: %s): %v", notification.DeviceID, notification.Platform, err)
		metrics.RecordRequest(notification.Platform, false)
		writeJSONResponse(w, http.StatusOK, push.PushResponse{
			Status: "FAIL",
			Error:  err.Error(),
		})
		return
	}

	metrics.RecordRequest(notification.Platform, true)
	writeJSONResponse(w, http.StatusOK, push.PushResponse{
		Status: "OK",
	})
}

// HandleAck handles GET/POST /api/v1/ack ping requests from Mattermost server
func (h *HandlerContainer) HandleAck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"OK","version":"2.0.0"}`))
}

// HandleHealthz handles system health checks
func (h *HandlerContainer) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"HEALTHY","service":"mattermost-push-proxy"}`))
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, resp push.PushResponse) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
