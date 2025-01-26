// Package middleware provides utility functions and structures
// for processing HTTP requests and responses, including compression
// and decompression of data.
package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// compressWriter implements the http.ResponseWriter interface, allowing
// transparent compression of data sent to the client and setting the correct
// HTTP headers.
type compressWriter struct {
	w  http.ResponseWriter // original ResponseWriter
	zw *gzip.Writer        // gzip.Writer for compressing data
}

// compressReader implements the io.ReadCloser interface, enabling transparent
// decompression of data received from the client.
type compressReader struct {
	r  io.ReadCloser // original request body
	zr *gzip.Reader  // gzip.Reader for decompressing data
}

// gzipWriterPool provides a pool of gzip.Writer objects for reuse,
// reducing memory allocations and improving performance.
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// newCompressWriter creates a new compressWriter to compress HTTP responses.
// It uses the gzip.Writer pool to efficiently manage resources.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	writer := gzipWriterPool.Get().(*gzip.Writer)
	writer.Reset(w)
	return &compressWriter{w: w, zw: writer}
}

// newCompressReader creates a new compressReader to decompress HTTP requests.
// If creating a gzip.Reader fails, an error is returned.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: zr}, nil
}

// Header returns the HTTP headers of the response.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write compresses and writes data to the underlying ResponseWriter.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes the HTTP status code to the response
// and sets the Content-Encoding header for successful (2xx) responses.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close completes the compression process, closes the gzip.Writer,
// and returns it to the pool for reuse.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw)
	return err
}

// Read reads decompressed data from the gzip.Reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes the compressReader, ensuring both the original body
// and the gzip.Reader are properly closed.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// CompressMiddleware is an HTTP middleware that automatically compresses
// HTTP responses if the client supports gzip and decompresses HTTP requests
// if the client sends gzipped data.
//
// If the client includes `Content-Encoding: gzip` in the request header,
// the middleware decompresses the request body. If the client includes
// `Accept-Encoding: gzip` in the request header, the middleware compresses
// the response body.
func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle gzip decoding for incoming requests
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			compressReader, err := newCompressReader(r.Body)
			if err != nil {
				log.Printf("Error decompressing request body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer compressReader.Close()
			r.Body = compressReader
		}

		// Handle gzip encoding for outgoing responses
		originalWriter := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			compressWriter := newCompressWriter(w)
			defer compressWriter.Close()
			originalWriter = compressWriter
		}

		// Pass control to the next handler in the chain
		h.ServeHTTP(originalWriter, r)
	})
}
