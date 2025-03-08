package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	// store program (bundled version)
	http.HandleFunc("PUT /store", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, r.ContentLength)
		r.Body.Read(data)

		fmt.Println(string(data))
	})

	// run program (sha256 hash)
	http.HandleFunc("POST /{hash}", func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")

		// ensure hash is valid decodable hex
		if len(hash) != 64 {
			http.Error(w, "Invalid hash length", http.StatusBadRequest)
			return
		} else if strings.ToLower(hash) != hash {
			http.Error(w, "Invalid hash case", http.StatusBadRequest)
			return
		} else if _, err := hex.DecodeString(hash); err != nil {
			http.Error(w, "Invalid hash", http.StatusBadRequest)
			return
		}

		// yeah
		fmt.Println(hash)
	})

	panic(http.ListenAndServe(":2505", nil))
}
