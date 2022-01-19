package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	//"time"

	"github.com/gorilla/mux"
)

const staticLocation = "/home/zzysonny/Code/Project/nTopia/Core/static/"

var initData sync.Map
var m4sCache = *CreateWindowedCache()

//receive mpd and initial data from ffmpeg
func ffmpegStaticHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FFMPEG]\t[MPD]\t[Start]\t %s\n", r.URL.Path)
	//read manifest data from body and overwrite
	data, _ := io.ReadAll(r.Body)
	initData.Store(r.URL.Path, data)
	log.Printf("[FFMPEG]\t[MPD]\t[Done]\t %s\n", r.URL.Path)
}

//receive m4s from ffmpeg
func ffmpegChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FFMPEG]\t[Chunk]\t[Start]\t  %s\n", r.URL.Path)
	cache := NewCachedWriter()
	m4sCache.CacheAdd(r.URL.Path, cache)

	_, err := io.Copy(cache, r.Body)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		cache.Close()
		log.Printf("[FFMPEG]\t[Chunk]\t[Done]\t %s\n", r.URL.Path)
	}
}

//serve mpd and inital data as dash server
func dashStaticHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Dash]\t[MPD]\t[Start]\t %s\n", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	data, _ := initData.Load(r.URL.Path)
	io.Copy(w, bytes.NewBuffer(data.([]byte)))
	log.Printf("[Dash]\t[MPD]\t[Done]\t %s\n", r.URL.Path)
}

//serve m4s as dash server
func dashChunkHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Dash]\t[Chunk]\t[Start]\t %s\n", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	cache := m4sCache.CacheFind(r.URL.Path)
	if cache == nil {
		return
	}

	cr := cache.CreateReader()
	io.Copy(w, cr)
	log.Printf("[Dash]\t[Chunk]\t[Done]\t %s\n", r.URL.Path)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ldash/stream.mpd", ffmpegStaticHandler).Methods("PUT", "POST")
	r.HandleFunc("/ldash/stream.mpd", dashStaticHandler).Methods("GET")
	r.HandleFunc("/ldash/init-stream{id:[0-9]+}.m4s", ffmpegStaticHandler).Methods("PUT", "POST")
	r.HandleFunc("/ldash/init-stream{id:[0-9]+}.m4s", dashStaticHandler).Methods("GET")
	r.HandleFunc("/ldash/chunk-stream{id:[0-9]+}-{id:[0-9]+}.m4s", ffmpegChunkHandler).Methods("PUT", "POST")
	r.HandleFunc("/ldash/chunk-stream{id:[0-9]+}-{id:[0-9]+}.m4s", dashChunkHandler).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticLocation)))

	http.ListenAndServe(":8080", r)
}
