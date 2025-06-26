package main

import (
	"crypto/sha3"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Heliodex/coputer/bundle"
	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/litecode/vm"
)

const NamesDir = bundle.DataDir + "/names"

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

// public key doesn't really need to be decoded so
func checkPK(w http.ResponseWriter, hash string) (b bool) {
	if len(hash) != 49 {
		http.Error(w, "Invalid public key length", http.StatusBadRequest)
		return
	} else if strings.ToLower(hash) != hash {
		http.Error(w, "Invalid public key case", http.StatusBadRequest)
		return
	}

	return true
}

func findExists(w http.ResponseWriter, hexhash string) (b bool) {
	if !bundle.BundleStored(hexhash) {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	return true
}

func runWebHash(w http.ResponseWriter, r *http.Request, hexhash string, c Compiler, errCache map[[32]byte]map[[32]byte]error, runCache map[[32]byte]map[[32]byte]ProgramRets) {
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
	var args WebArgs
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
		runCache[hash] = make(map[[32]byte]ProgramRets, 1)
	}
	runCache[hash][inputhash] = output

	b, err := json.Marshal(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func main() {
	c := vm.NewCompiler(1)

	// just 1 error cache, as different inputs may result in errors/not
	// (we don't want one error to bring down the whole program for every user)
	errCache := make(map[[32]byte]map[[32]byte]error)
	runCache := make(map[[32]byte]map[[32]byte]ProgramRets)

	// store program (bundled version)
	http.HandleFunc("PUT /store/{pk}/{name}", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("put time")
		pk, name := r.PathValue("pk"), r.PathValue("name")
		if !checkPK(w, pk) {
			return
		}

		data := make([]byte, r.ContentLength)
		r.Body.Read(data)

		hash := sha3.Sum256(data)

		if err := os.MkdirAll(filepath.Join(NamesDir, pk), 0o755); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// fmt.Println("CREATING", filepath.Join(NamesDir, pk, name))
		f, err := os.Create(filepath.Join(NamesDir, pk, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest) // status whatever, reeks of ego anyway
			return
		}
		defer f.Close()

		if _, err := f.Write(hash[:]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Write to file after creating names paths
		if hexhash := hex.EncodeToString(hash[:]); bundle.BundleStored(hexhash) {
			http.Error(w, "Program already exists", http.StatusConflict)
			return
		} else if _, err := bundle.UnbundleToDir(data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	http.HandleFunc("GET /{pk}/{name}", func(w http.ResponseWriter, r *http.Request) {
		pk, name := r.PathValue("pk"), r.PathValue("name")
		if !checkPK(w, pk) {
			return
		}

		// fmt.Println("FINDING", filepath.Join(NamesDir, pk, name))
		hash, err := os.ReadFile(filepath.Join(NamesDir, pk, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		hexhash := hex.EncodeToString(hash)
		if _, ok := checkHash(w, hexhash); !ok {
			return
		}

		findExists(w, hexhash)
	})

	http.HandleFunc("POST /web/{pk}/{name}", func(w http.ResponseWriter, r *http.Request) {
		pk, name := r.PathValue("pk"), r.PathValue("name")
		if !checkPK(w, pk) {
			return
		}

		// fmt.Println("READING", filepath.Join(NamesDir, pk, name))
		hash, err := os.ReadFile(filepath.Join(NamesDir, pk, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		hexhash := hex.EncodeToString(hash)
		if _, ok := checkHash(w, hexhash); !ok {
			return
		}

		runWebHash(w, r, hexhash, c, errCache, runCache)
	})

	fmt.Println("Listening on :2505")
	panic(http.ListenAndServe(":2505", nil))
}
