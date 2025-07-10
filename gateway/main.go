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
	host    = "localhost"
	dothost = "." + host

	egPk = "1mdy2o0f9-s1a9rdjkt-vwut3s6fv-gd1nv0ezr-it04zc2le"

	commPort = 2507
	hostPort = 2517
)

func serveMain(w http.ResponseWriter, _ *http.Request, host string) {
	fmt.Fprintln(w, "Welcome to the Coputer Gateway!")
	fmt.Fprintln(w, "This is a placeholder for the main page.")
	fmt.Fprintf(w, "Profiles are accessible via subdomains like %s.%s\n", egPk, host)
	fmt.Fprintf(w, "Programs are accessible via subdomains like example.%s.%s\n", egPk, host)
}

func serveProfile(w http.ResponseWriter, _ *http.Request, pk keys.PK) {
	fmt.Fprintf(w, "Profile for public key %s\n", pk.Encode())
	fmt.Fprintln(w, "This is a placeholder for the public key's profile page.")
}

func serveWeb(w http.ResponseWriter, r *http.Request, pk keys.PK, name string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	url := WebArgsUrl{
		Rawpath:  r.URL.RawPath,
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

	if !strings.HasSuffix(hn, host) {
		http.Error(w, fmt.Sprintf("Invalid hostname: %s", hn), http.StatusBadRequest)
		return
	}

	sub := strings.TrimSuffix(hn, dothost)

	// subdomain is either {name}.{pk} or just {pk}
	dots := strings.Count(sub, ".")
	if dots > 1 {
		// subdomains of subdomains? someday
		http.Error(w, fmt.Sprintf("Invalid subdomain: %s", sub), http.StatusBadRequest)
		return
	}

	var name, pks string
	if dots == 0 {
		// subdomain is just {pk}
		pks = sub
	} else {
		// subdomain is {name}.{pk}
		parts := strings.SplitN(sub, ".", 2)
		name, pks = parts[0], parts[1]
	}

	pk, err := keys.DecodePKNoPrefix(pks)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode public key: %v", err), http.StatusBadRequest)
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
	fmt.Println("Starting")

	// match any route as a subdomain of localhost
	http.HandleFunc("/", handleRoute)

	fmt.Println("Hosting on port", hostPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", hostPort), nil); err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
}
