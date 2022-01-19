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
	//bytes received so far
	data   []byte
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

//Create a reader that reads all written bytes afresh
func (w *CachedWriter) CreateReader() io.Reader {
	return &CachedReader{
		source:   w,
		lastRead: 0,
	}
}

func (r *CachedReader) Read(p []byte) (n int, err error) {
	w := *r.source
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
type ChunkWindowedCache struct{
	mutex sync.Mutex
	cur int
	k[windowsize] string
	v[windowsize] *CachedWriter
}

func CreateWindowedCache() *ChunkWindowedCache {
	return &ChunkWindowedCache{}
}

func (c *ChunkWindowedCache) CacheAdd(url string, cache *CachedWriter){
	c.mutex.Lock()
	c.k[c.cur]=url
	c.v[c.cur]=cache
	c.cur=(c.cur+1)%windowsize
	c.mutex.Unlock()
}

func (c *ChunkWindowedCache) CacheFind(url string) *CachedWriter{
	c.mutex.Lock()
	for i:=0; i<windowsize; i++ {
		if c.k[i]==url {
			ans:=c.v[i]
			c.mutex.Unlock()
			return ans
		}
	}
	c.mutex.Unlock()
	return nil
}