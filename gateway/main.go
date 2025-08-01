package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const (
	egPk = "1mdy2o0f9-s1a9rdjkt-vwut3s6fv-gd1nv0ezr-it04zc2le"

	commPort = 2507
	hostPort = 2517
)

var host = "localhost"

func validateSubdomain(hn string) (pk keys.PK, name string, err error) {
	if !strings.HasSuffix(hn, host) {
		return keys.PK{}, "", fmt.Errorf("invalid hostname: %s", hn)
	}

	sub := strings.TrimSuffix(hn, "."+host)
	ls := len(sub)

	// subdomain is either {name}-{pk} or just {pk}
	if ls != 49 && ls <= 50 {
		return keys.PK{}, "", fmt.Errorf("invalid subdomain: %s", sub)
	}

	pks := sub[ls-49:] // last 49 characters are the public key
	if ls > 50 {
		if sub[ls-50] != '-' {
			return keys.PK{}, "", fmt.Errorf("invalid subdomain format: %s", sub)
		}
		name = sub[:ls-50]
	}

	pk, err = keys.DecodePKNoPrefix(pks)
	if err != nil {
		return keys.PK{}, "", fmt.Errorf("failed to decode public key: %v", err)
	}
	return
}

func serveMain(w http.ResponseWriter, _ *http.Request, host string) {
	fmt.Fprintln(w, "Welcome to the Coputer Gateway!")
	fmt.Fprintln(w, "This is a placeholder for the main page.")
	fmt.Fprintf(w, "Profiles are accessible via subdomains like %s.%s\n", egPk, host)
	fmt.Fprintf(w, "Programs are accessible via subdomains like example-%s.%s\n", egPk, host)
}

func serveProfile(w http.ResponseWriter, r *http.Request, pk keys.PK) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	programs, err := GetProfile(pk)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get profile: %v", err), http.StatusInternalServerError)
		return
	}

	if len(programs) == 0 {
		fmt.Fprintf(w, "No programs found for public key %s\n", pk.Encode())
		return
	}

	fmt.Fprintf(w, "Programs for public key %s:\n", pk.Encode())
	for _, prog := range programs {
		fmt.Fprintf(w, "- %s-%s.%s\n", prog, pk.EncodeNoPrefix(), host)
	}
}

func serveWeb(w http.ResponseWriter, r *http.Request, pk keys.PK, name string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	url := WebArgsUrl{
		Rawpath:  r.URL.String(),
		Path:     r.URL.Path, // always absolute, I think
		Rawquery: r.URL.RawQuery,
		Query:    r.URL.Query(),
	}

	headers := make(map[string]string, len(r.Header))
	for k := range r.Header {
		headers[k] = r.Header.Get(k)
	}

	args := WebArgs{
		Url:     url,
		Method:  r.Method,
		Headers: headers,
		Body:    body,
	}

	rets, err := StartWebProgram(pk, name, args)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start web program: %v", err), http.StatusInternalServerError)
		return
	}

	for k, v := range rets.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(rets.StatusCode)
	w.Write(rets.Body)
}

func handleRoute(w http.ResponseWriter, r *http.Request) {
	u, err := r.URL.Parse("https://" + r.Host)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	hn := u.Hostname()
	if hn == host {
		// serve main page
		serveMain(w, r, host)
		return
	}

	pk, name, err := validateSubdomain(hn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if name == "" {
		// show public key "profile"
		serveProfile(w, r, pk)
		return
	}

	// serve web program
	serveWeb(w, r, pk, name)
}

func main() {
	if len(os.Args) > 1 {
		host = os.Args[1]
	}

	fmt.Println("Starting")

	// match any route as a subdomain of localhost
	http.HandleFunc("/", handleRoute)

	fmt.Println("Hosting on port", hostPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", hostPort), nil); err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
}
