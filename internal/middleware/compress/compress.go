package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

// func newCompressWriter(w http.ResponseWriter) (*compressWriter, error) {
// 	zw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
// 	if err != nil {
// 		return nil, err
// 	}

//		return &compressWriter{
//			w:  w,
//			zw: zw,
//		}, nil
//	}
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	writer := gzipWriterPool.Get().(*gzip.Writer)
	writer.Reset(w)
	return &compressWriter{w: w, zw: writer}
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// type sizeCheckingWriter struct {
// 	http.ResponseWriter
// 	Buffer io.Writer
// 	Size   int
// }

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// func (w *sizeCheckingWriter) Write(p []byte) (int, error) {
// 	w.Size += len(p)
// 	if w.Size > 1024 { // минимум 1KB
// 		return w.Buffer.Write(p)
// 	}
// 	return len(p), nil
// }

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw) // Возвращаем writer в пул
	return err
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		// originalWriter := w
		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			compressReader, err := newCompressReader(r.Body)
			if err != nil {
				log.Printf("Ошибка при декомпрессии: %v", err)

				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			defer compressReader.Close()
			r.Body = compressReader
		}
		originalWriter := w
		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			compressWriter := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			defer compressWriter.Close()
			originalWriter = compressWriter
			// не забываем отправить клиенту все сжатые данные после завершения middleware
		}
		// передаём управление хендлеру
		h.ServeHTTP(originalWriter, r)
	})
}
