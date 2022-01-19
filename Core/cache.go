package main

import (
	"io"
	"sync"
)

//CacheWriter implements io.WriteCloser
//
//The writer allows to create a reader that reads
//all past bytes and future bytes.
type CachedWriter struct {
	//mutex
	//mutex sync.Mutex
	//bytes received so far
	data []byte
	//if the writer is closed
	closed bool
}

//CachedReader implements io.Reader
//
//Read byte from source CachedWriter until
//source is closed and all bytes are read
type CachedReader struct {
	source   *CachedWriter
	lastRead int
}

//Creat new CachedWriter
func NewCachedWriter() *CachedWriter {
	return &CachedWriter{
		data:   []byte{},
		closed: false,
	}
}

func (w *CachedWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	//all bytes are successfully written
	return len(p), nil
}

func (w *CachedWriter) Close() error {
	w.closed = true
	return nil
}

func (w *CachedWriter) Clear() {
	w.data = w.data[:0]
	w.closed = false
}

//Create a reader that reads all written bytes afresh
func (w *CachedWriter) CreateReader() io.Reader {
	return &CachedReader{
		source:   w,
		lastRead: 0,
	}
}

func (r *CachedReader) Read(p []byte) (n int, err error) {
	w := r.source
	//log.Println(w.closed, len(w.data), r.lastRead)
	//if no more bytes will come and all bytes is read
	if w.closed && len(w.data) == r.lastRead {
		//EOF
		if len(p) == 0 {
			return 0, nil
		} else {
			return 0, io.EOF
		}
	}
	//otherwise read some bytes
	n = copy(p, w.data[r.lastRead:])
	r.lastRead += n
	return n, nil
}

const windowsize = 30

//Windowed cache
type ChunkWindowedCache struct {
	mutex  sync.Mutex
	cur    int
	path   [windowsize]string
	writer [windowsize]CachedWriter
}

func CreateWindowedCache() *ChunkWindowedCache {
	return &ChunkWindowedCache{}
}

func (c *ChunkWindowedCache) Allocate(url string) *CachedWriter {
	c.mutex.Lock()

	c.path[c.cur] = url
	ans := &c.writer[c.cur]
	ans.Clear()
	c.cur = (c.cur + 1) % windowsize

	c.mutex.Unlock()
	return ans
}

func (c *ChunkWindowedCache) Find(url string) *CachedWriter {
	c.mutex.Lock()
	for i := 0; i < windowsize; i++ {
		if c.path[i] == url {
			ans := &c.writer[i]
			c.mutex.Unlock()
			return ans
		}
	}
	c.mutex.Unlock()
	return nil
}

func (c *ChunkWindowedCache) Restart()  {
	c.cur=0
}
