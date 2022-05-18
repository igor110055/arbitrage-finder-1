package middlewares

import (
	"bufio"
	"calc/foundation/id"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Logger.
			With().
			Stringer("RequestID", id.ULID()).
			Logger()

		ctx := logger.WithContext(r.Context())

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				cause := errors.WithStack(err.(error))
				logger.Error().
					Stack().
					Err(cause).
					Msg("log middleware")
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r.WithContext(ctx))
		logger.Info().Msgf("status: %d; method: %s; path: %s; duration: %d\n", wrapped.status, r.Method, r.URL.EscapedPath(), time.Since(start))
	})
}
