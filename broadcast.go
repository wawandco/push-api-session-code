package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// NotificationPayload represents the notification data
type NotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Badge string `json:"badge,omitempty"`
	Data  any    `json:"data,omitempty"`
}

// Broadcast sends a notification to all subscribers
func (s *WebPushServer) Broadcast(w http.ResponseWriter, r *http.Request) {
	var payload NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	subscriptions := make([]*webpush.Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		subscriptions = append(subscriptions, sub)
	}
	s.mu.RUnlock()

	successCount := 0
	failCount := 0

	// Send to all subscriptions
	for _, sub := range subscriptions {
		if err := s.SendNotification(sub, payload); err != nil {
			log.Printf("Failed to send to %s: %v\n", sub.Endpoint[:50]+"...", err)
			failCount++
		} else {
			successCount++
		}
	}

	log.Printf("Broadcast complete: %d sent, %d failed\n", successCount, failCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"sent":    successCount,
		"failed":  failCount,
		"message": fmt.Sprintf("Notification sent to %d subscribers", successCount),
	})
}

// SendNotification sends a notification to a specific subscription
func (s *WebPushServer) SendNotification(subscription *webpush.Subscription, payload NotificationPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := webpush.SendNotification(jsonPayload, subscription, &webpush.Options{
		Subscriber:      "admin@example.com",
		VAPIDPublicKey:  s.vapidPublic,
		VAPIDPrivateKey: s.vapidPrivate,
		TTL:             30,
	})
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 410 || resp.StatusCode == 404 {
		// Subscription expired, remove it
		s.mu.Lock()
		delete(s.subscriptions, subscription.Endpoint)
		s.mu.Unlock()
		return fmt.Errorf("subscription expired (status %d)", resp.StatusCode)
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
