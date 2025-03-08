package main

import (
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/Heliodex/coputer/exec"
	"github.com/Heliodex/coputer/litecode/vm"
)

// ensure hash is valid decodable hex
func checkHash(w http.ResponseWriter, hash string) (b bool) {
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

	return true
}

func main() {
	c := vm.NewCompiler(1)

	retsCache := make(map[string]string)
	errCache := make(map[string]error)

	// store program (bundled version)
	http.HandleFunc("PUT /store", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, r.ContentLength)
		r.Body.Read(data)

		hash := sha3.Sum256(data)
		if hexhash := hex.EncodeToString(hash[:]); exec.BundleStored(hexhash) {
			return
		} else if _, err := exec.UnbundleToDir(data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	// find if program exists
	http.HandleFunc("POST /{hash}", func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")

		if !checkHash(w, hash) {
			return
		} else if !exec.BundleStored(hash) {
			http.Error(w, "Program not found", http.StatusNotFound)
			return
		}

		run := r.URL.Query().Has("run")
		if !run {
			return
		} else if res, ok := retsCache[hash]; ok {
			fmt.Fprintln(w, vm.ToString(res))
			return
		} else if err, ok := errCache[hash]; ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// run program
		res, err := Run(c, hash)
		if err != nil {
			errCache[hash] = err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		final := vm.ToString(res)
		retsCache[hash] = final
		fmt.Fprintln(w, final)
	})

	fmt.Println("Listening on :2505")
	panic(http.ListenAndServe(":2505", nil))
}
