package main

import (
	"crypto/sha3"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Heliodex/coputer/bundle"
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
	if !bundle.BundleStored(hash) {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	return true
}

func main() {
	c := vm.NewCompiler(1)

	// just 1 error cache, as different inputs may result in errors/not
	// (we don't want one error to bring down the whole program for every user)
	errCache := make(map[[32]byte]map[[32]byte]error)
	runCache := make(map[[32]byte]map[[32]byte]vm.ProgramRets)

	// store program (bundled version)
	http.HandleFunc("PUT /store", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, r.ContentLength)
		r.Body.Read(data)

		hash := sha3.Sum256(data)
		if hexhash := hex.EncodeToString(hash[:]); bundle.BundleStored(hexhash) {
			return
		} else if _, err := bundle.UnbundleToDir(data); err != nil {
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
	http.HandleFunc("POST /web/{hash}", func(w http.ResponseWriter, r *http.Request) {
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
		inputhash := sha3.Sum256(input) // let's hope it's canonical

		// decode input as json
		var args vm.WebArgs
		if err := json.Unmarshal(input, &args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !findExists(w, hexhash) {
			return
		} else if err, ok := errCache[hash][inputhash]; ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if res, ok := runCache[hash][inputhash]; ok {
			b, err := json.Marshal(res)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(b)
			return
		}

		// rets program
		output, err := Start(c, hexhash, args)
		if err != nil {
			if errCache[hash] == nil {
				errCache[hash] = make(map[[32]byte]error, 1)
			}
			errCache[hash][inputhash] = err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if runCache[hash] == nil {
			runCache[hash] = make(map[[32]byte]vm.ProgramRets, 1)
		}
		runCache[hash][inputhash] = output

		b, err := json.Marshal(output)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})

	fmt.Println("Listening on :2505")
	panic(http.ListenAndServe(":2505", nil))
}
