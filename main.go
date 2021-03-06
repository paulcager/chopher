package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pointlander/chopher/api"
	"github.com/pointlander/chopher/hasher"
	"github.com/pointlander/chopher/karplus"
	"github.com/pointlander/chopher/wave"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var file = flag.String("file", "", "file to hash")
var seed = flag.Int64("seed", 0, "random seed for song")

type RandReader struct {
	size int
	rnd  *rand.Rand
}

func NewRandReader(size int, seed int64) *RandReader {
	return &RandReader{
		size: size,
		rnd:  rand.New(rand.NewSource(seed)),
	}
}

func (rr *RandReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(rr.rnd.Int())
		rr.size--
		n++
		if rr.size == 0 {
			err = errors.New("End of random stream")
			return
		}
	}
	return
}

func main() {
	flag.Parse()

	if *file != "" {
		in, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}
		h := hasher.New(in)
		sng := h.Hash()
		in.Close()

		wav := wave.New(wave.Stereo, 22000)
		ks := karplus.Song{
			Song:         *sng,
			SamplingRate: 22000,
		}
		ks.Sound(&wav)

		out, err := os.Create(strings.TrimSuffix(*file, filepath.Ext(*file)) + ".wav")
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(out, wav.Reader())
		return
	}

	if *seed != 0 {
		in := NewRandReader(2*1024*1024, *seed)
		h := hasher.New(in)
		sng := h.Hash()

		wav := wave.New(wave.Stereo, 22000)
		ks := karplus.Song{
			Song:         *sng,
			SamplingRate: 22000,
		}
		ks.Sound(&wav)

		out, err := os.Create(fmt.Sprintf("%v.wav", *seed))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(out, wav.Reader())
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/"))).Methods("GET")
	r.HandleFunc("/upload", api.FileUploadHandler).Methods("POST")
	http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, r))
}
