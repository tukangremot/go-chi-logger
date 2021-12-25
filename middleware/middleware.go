package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

func Logger(category string, logger logrus.FieldLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var (
				reqID 	  = middleware.GetReqID(r.Context())
				ww    	  = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
				ts    	  = time.Now().UTC()
				host  	  = r.Host
				uri    	  = r.RequestURI
				userAgent = r.UserAgent()
			)

			defer func() {
				var (
					remoteIP, _, err = net.SplitHostPort(r.RemoteAddr)
					scheme           = "http"
					method           = r.Method
					duration         = time.Since(ts)
					addr             = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
				)

				if err != nil {
					remoteIP = r.RemoteAddr
				}
				if r.TLS != nil {
					scheme = "https"
				}
				fields := logrus.Fields{
					"http_host":         host,
					"http_uri":          uri,
					"http_proto":        r.Proto,
					"http_method":       method,
					"http_scheme":       scheme,
					"http_addr":         addr,
					"remote_addr":       remoteIP,
					"user_agent":	     userAgent,
					"resp_status":       ww.Status(),
					"resp_elapsed":      int64(duration),
					"resp_elapsed_ms":   time.Since(ts).String(),
					"resp_bytes_length": ww.BytesWritten(),
					"ts":                ts.Format(time.RFC1123),
					"category":          category,
				}
				if len(reqID) > 0 {
					fields["request_id"] = reqID
				}
				logger.WithFields(fields).Infof("[%s] %s://%s%s", method, scheme, host, uri)
			}()

			h.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
