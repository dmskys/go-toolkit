package next

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mylxsw/go-toolkit/log"
)

// CreateHTTPHandler create a http handler for request processing
func CreateHTTPHandler(config *Config) http.Handler {
	rootDir := filepath.Dir(config.EndpointFile)

	handler := Handler{
		Rules:           config.Rules,
		Root:            rootDir,
		FileSys:         http.Dir(rootDir),
		SoftwareName:    config.SoftwareName,
		SoftwareVersion: config.SoftwareVersion,
		ServerName:      config.ServerIP,
		ServerPort:      strconv.Itoa(config.ServerPort),
	}

	return &HTTPHandler{
		handler: handler,
		config:  config,
	}
}

// RequestLogHandler request log handler func
type RequestLogHandler func(rc *RequestContext)

// Config config object for create a handler
type Config struct {
	EndpointFile      string
	ServerIP          string
	ServerPort        int
	SoftwareName      string
	SoftwareVersion   string
	RequestLogHandler RequestLogHandler
	Rules             []Rule
}

// HTTPHandler http request handler wrapper
type HTTPHandler struct {
	handler Handler
	config  *Config
}

// RequestContext requext context information
type RequestContext struct {
	UA      string
	Method  string
	Referer string
	Headers http.Header
	URI     string
	Body    url.Values
	Consume time.Duration
	Code    int
}

// ToMap convert the requestContext to a map
func (rc *RequestContext) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ua":      rc.UA,
		"method":  rc.Method,
		"referer": rc.Referer,
		"headers": rc.Headers,
		"uri":     rc.URI,
		"body":    rc.Body,
		"consume": fmt.Sprintf("%.4f", rc.Consume.Seconds()),
		"code":    rc.Code,
	}
}

// ServeHTTP implements http.Handler interface
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var statusCode = 200
	respWriter := NewResponseWriter(w, func(code int) {
		statusCode = code
	})
	defer func(startTime time.Time) {
		consume := time.Now().Sub(startTime)
		if r.Form == nil {
			r.ParseForm()
		}

		if h.config.RequestLogHandler != nil {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Module("next").Errorf("request log handler has some error: %v", err)
					}
				}()

				h.config.RequestLogHandler(&RequestContext{
					UA:      r.UserAgent(),
					Method:  r.Method,
					Referer: r.Referer(),
					Headers: r.Header,
					URI:     r.RequestURI,
					Body:    r.Form,
					Consume: consume,
					Code:    statusCode,
				})
			}()
		}

	}(time.Now())

	code, err := h.handler.ServeHTTP(respWriter, r)
	if err != nil {
		log.Module("next").Errorf("request failed, code=%d, err=%s", code, err.Error())
	}
}
