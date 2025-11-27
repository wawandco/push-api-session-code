package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// Subscription represents a push subscription
type Subscription struct {
	Endpoint       string `json:"endpoint"`
	ExpirationTime *int64 `json:"expirationTime"`
	Keys           struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// WebPushServer handles push notifications
type WebPushServer struct {
	subscriptions map[string]*webpush.Subscription
	mu            sync.RWMutex
	vapidPublic   string
	vapidPrivate  string
}

// NewWebPushServer creates a new push server
func NewWebPushServer(vapidPublic, vapidPrivate string) *WebPushServer {
	return &WebPushServer{
		subscriptions: make(map[string]*webpush.Subscription),
		vapidPublic:   vapidPublic,
		vapidPrivate:  vapidPrivate,
	}
}

// Subscribe handles subscription requests
func (s *WebPushServer) Subscribe(w http.ResponseWriter, r *http.Request) {
	var sub Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to webpush.Subscription
	wpSub := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.Keys.P256dh,
			Auth:   sub.Keys.Auth,
		},
	}

	// Store subscription (using endpoint as key)
	s.mu.Lock()
	s.subscriptions[sub.Endpoint] = wpSub
	s.mu.Unlock()

	log.Printf("New subscription added: %s\n", sub.Endpoint[:50]+"...")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Subscription saved",
	})
}

// Unsubscribe handles unsubscribe requests
func (s *WebPushServer) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	delete(s.subscriptions, req.Endpoint)
	s.mu.Unlock()

	log.Printf("Subscription removed: %s\n", req.Endpoint[:50]+"...")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Unsubscribed",
	})
}

// GetStats returns server statistics
func (s *WebPushServer) GetStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	count := len(s.subscriptions)
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"totalSubscriptions": count,
	})
}

// GetVAPIDPublicKey returns the VAPID public key
func (s *WebPushServer) GetVAPIDPublicKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"publicKey": s.vapidPublic,
	})
}
