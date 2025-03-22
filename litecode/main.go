package main

import (
	"crypto/sha3"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Heliodex/coputer/exec"
	"github.com/Heliodex/coputer/litecode/vm"
)

// ensure hash is valid decodable hex
func checkHash(w http.ResponseWriter, hash string) (decoded [32]byte, b bool) {
	if len(hash) != 64 {
		http.Error(w, "Invalid hash length", http.StatusBadRequest)
		return
	} else if strings.ToLower(hash) != hash {
		http.Error(w, "Invalid hash case", http.StatusBadRequest)
		return
	}

	dechex, err := hex.DecodeString(hash)
	if err != nil {
		http.Error(w, "Invalid hash", http.StatusBadRequest)
		return
	}

	return [32]byte(dechex), true
}

func findExists(w http.ResponseWriter, hash string) (b bool) {
	if !exec.BundleStored(hash) {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	return true
}

func main() {
	c := vm.NewCompiler(1)

	startErrCache := make(map[[32]byte]error)
	inputErrCache := make(map[[32]byte]map[[32]byte]error)
	runCache := make(map[[32]byte]map[[32]byte]string)

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
	http.HandleFunc("GET /{hash}", func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")
		if _, ok := checkHash(w, hash); !ok {
			return
		}

		findExists(w, hash)
	})

	// run program
	http.HandleFunc("POST /{hash}", func(w http.ResponseWriter, r *http.Request) {
		hexhash := r.PathValue("hash")
		hash, ok := checkHash(w, hexhash)
		if !ok {
			return
		}

		input, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		inputhash := sha3.Sum256(input)

		if !findExists(w, hexhash) {
			return
		} else if err, ok := startErrCache[hash]; ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err, ok := inputErrCache[hash][inputhash]; ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if res, ok := runCache[hash][inputhash]; ok {
			fmt.Fprintln(w, vm.ToString(res))
			return
		}

		// run program
		run, err := Start(c, hexhash)
		if err != nil {
			startErrCache[hash] = err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		output, err := run(string(input))
		if err != nil {
			if _, ok := inputErrCache[hash]; !ok {
				inputErrCache[hash] = make(map[[32]byte]error)
			}
			inputErrCache[hash][inputhash] = err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if _, ok := runCache[hash]; !ok {
			runCache[hash] = make(map[[32]byte]string)
		}
		runCache[hash][inputhash] = output
		fmt.Fprintln(w, output)
	})

	fmt.Println("Listening on :2505")
	panic(http.ListenAndServe(":2505", nil))
}
