package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const staticLocation = "/home/zzysonny/Code/Project/nTopia/Core/static/"

var mpddata []byte
var initdata []byte
var m4sCache = *CreateWindowedCache()

//receive mpd data from ffmpeg
func ffmpegMPDReceiver(w http.ResponseWriter, r *http.Request) {
	//read manifest data from body and overwrite
	mpddata, _ = io.ReadAll(r.Body)
	log.Printf("[FFMPEG]\t[MPD]\t[Done]\t %s\n", r.URL.Path)
}

//receive initial data from ffmpeg
func ffmpegInitReceiver(w http.ResponseWriter, r *http.Request) {
	//read init data from body
	m4sCache.Restart()
	initdata, _ = io.ReadAll(r.Body)
	log.Printf("[FFMPEG]\t[Init]\t[Done]\t %s\n", r.URL.Path)
}

//receive m4s from ffmpeg
func ffmpegChunkReceiver(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FFMPEG]\t[Chunk]\t[Start]\t %s\n", r.URL.Path)
	cache := m4sCache.Allocate(r.URL.Path)
	len, err := io.Copy(cache, r.Body)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		cache.Close()
		log.Printf("[FFMPEG]\t[Chunk]\t[Done]\t %s %d\n", r.URL.Path, len)
	}
}

//serve mpd data as dash server
func dashMPDSender(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Allow-Headers", "Content-Type")
	r.Header.Set("Connection", "Keep-Alive")
	r.Header.Set("Content-Type", "application/dash+xml")
	io.Copy(w, bytes.NewBuffer(mpddata))
	log.Printf("[Dash]\t[MPD]\t[Done]\t %s\n", r.URL.Path)
}

//serve init data as dash server
func dashInitSender(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Allow-Headers", "Content-Type")
	r.Header.Set("Connection", "Keep-Alive")
	r.Header.Set("Content-Type", "video/iso.segment")
	io.Copy(w, bytes.NewBuffer(initdata))
	log.Printf("[Dash]\t[MPD]\t[Done]\t %s\n", r.URL.Path)
}

//serve m4s as dash server
func dashChunkSender(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Dash]\t[Chunk]\t[Start]\t %s\n", r.URL.Path)
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Allow-Headers", "Content-Type")
	r.Header.Set("Transfer-Encoding", "chunked")
	r.Header.Set("Connection", "Keep-Alive")
	r.Header.Set("Content-Type", "video/iso.segment")

	cache := m4sCache.Find(r.URL.Path)
	if cache == nil {
		return
	}
	//httputil.NewChunkedWriter()

	cr := cache.CreateReader()
	len, _ := io.Copy(w, cr)
	log.Printf("[Dash]\t[Chunk]\t[Done]\t %s %d\n", r.URL.Path, len)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ldash/stream.mpd", ffmpegMPDReceiver).Methods("PUT", "POST")
	r.HandleFunc("/ldash/stream.mpd", dashMPDSender).Methods("GET")
	r.HandleFunc("/ldash/init-stream{id:[0-9]+}.m4s", ffmpegInitReceiver).Methods("PUT", "POST")
	r.HandleFunc("/ldash/init-stream{id:[0-9]+}.m4s", dashInitSender).Methods("GET")
	r.HandleFunc("/ldash/chunk-stream{id:[0-9]+}-{id:[0-9]+}.m4s", ffmpegChunkReceiver).Methods("PUT", "POST")
	r.HandleFunc("/ldash/chunk-stream{id:[0-9]+}-{id:[0-9]+}.m4s", dashChunkSender).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticLocation)))

	http.ListenAndServe(":8080", r)
}
