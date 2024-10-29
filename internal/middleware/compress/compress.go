package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
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

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
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

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
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
		originalWriter := w
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
