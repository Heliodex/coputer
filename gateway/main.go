package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	host = "localhost"
	dothost = "." + host
)

func main() {
	fmt.Println("Starting")

	// match any route as a subdomain of localhost
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u, err := r.URL.Parse("https://" + r.Host)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		hn := u.Hostname()
		if hn == host {
			// serve main page
			fmt.Fprintln(w, "Hello from main page")
			return
		}

		if !strings.HasSuffix(hn, host) {
			http.Error(w, fmt.Sprintf("Invalid hostname: %s", hn), http.StatusBadRequest)
			return
		}

		sub := strings.TrimSuffix(hn, dothost)
		if strings.Contains(sub, ".") {
			http.Error(w, fmt.Sprintf("Invalid subdomain: %s", sub), http.StatusBadRequest)
			return
		}

		pk, err := keys.DecodePK(sub)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode public key: %v", err), http.StatusBadRequest)
			return
		}

		// Respond with a simple message
		fmt.Fprintln(w, "Hello from public key:", pk)
	})

	fmt.Println("Listening on port 2507")
	panic(http.ListenAndServe(":2507", nil))
}
