package main

import (
	"fmt"
	"log"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// GenerateVAPIDKeys generates new VAPID keys
func GenerateVAPIDKeys() (privateKey, publicKey string, err error) {
	return webpush.GenerateVAPIDKeys()
}

func main() {
	// Generate VAPID keys (you should do this once and save them)
	privateKey, publicKey, err := GenerateVAPIDKeys()
	if err != nil {
		log.Fatal("Failed to generate VAPID keys:", err)
	}

	fmt.Println("=== VAPID Keys Generated ===")
	fmt.Println("Public Key:", publicKey)
	fmt.Println("Private Key:", privateKey)
	fmt.Println("Save these keys for production use!")

	// Create server
	server := NewWebPushServer(publicKey, privateKey)

	// Setup routes
	http.HandleFunc("POST /subscribe", server.Subscribe)
	http.HandleFunc("POST /unsubscribe", server.Unsubscribe)
	http.HandleFunc("POST /broadcast", server.Broadcast)
	http.HandleFunc("GET /vapid-public-key", server.GetVAPIDPublicKey)
	http.HandleFunc("GET /stats", server.GetStats)

	// Serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, "index.html")
		case "/service-worker.js":
			w.Header().Set("Content-Type", "application/javascript")
			http.ServeFile(w, r, "service-worker.js")
		default:
			http.NotFound(w, r)
		}
	})

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	fmt.Println("Make sure to create index.html and service-worker.js in the same directory")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
