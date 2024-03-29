package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/BranLwyd/harpocrates/harpd/assets"
)

func serveTemplate(w http.ResponseWriter, r *http.Request, tmpl *template.Template, data interface{}) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Could not execute %q template: %v", tmpl.Name(), err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	newStatic(buf.Bytes(), "text/html; charset=utf-8").ServeHTTP(w, r)
}

func must(h http.Handler, err error) http.Handler {
	if err != nil {
		panic(err)
	}
	return h
}

// staticHandler serves static content from memory.
type staticHandler struct {
	content     []byte
	contentType string
}

func newStatic(content []byte, contentType string) staticHandler {
	return staticHandler{
		content:     content,
		contentType: contentType,
	}
}

func newAsset(name, contentType string) (staticHandler, error) {
	asset, ok := assets.Asset[name]
	if !ok {
		return staticHandler{}, fmt.Errorf("no such asset %q", name)
	}
	return newStatic(asset, contentType), nil
}

func (sh staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", sh.contentType)
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(sh.content))
}

type cacheableStaticHandler struct {
	sh staticHandler

	tagOnce sync.Once
	tag     string
}

func newCacheableAsset(name, contentType string) (*cacheableStaticHandler, error) {
	sh, err := newAsset(name, contentType)
	if err != nil {
		return nil, err
	}
	csh := &cacheableStaticHandler{sh: sh}
	go csh.etag() // eagerly compute etag so that it will probably be available by the first request
	return csh, nil
}

func (csh cacheableStaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", csh.etag())
	csh.sh.ServeHTTP(w, r)
}

func (csh cacheableStaticHandler) etag() string {
	csh.tagOnce.Do(func() {
		h := sha256.Sum256(csh.sh.content)
		csh.tag = fmt.Sprintf(`"%s"`, base64.RawURLEncoding.EncodeToString(h[:]))
	})
	return csh.tag
}

// secureHeaderHandler adds a few security-oriented headers.
type secureHeaderHandler struct {
	h http.Handler
}

func (shh secureHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")

	shh.h.ServeHTTP(w, r)
}

func NewSecureHeader(h http.Handler) http.Handler {
	return secureHeaderHandler{h}
}

// filteredHandler filters a handler to only serve one path; anything else is given a 404.
type filteredHandler struct {
	allowedPath string
	h           http.Handler
}

func newFiltered(allowedPath string, h http.Handler) http.Handler {
	return &filteredHandler{
		allowedPath: allowedPath,
		h:           h,
	}
}

func (fh filteredHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Clean(r.URL.Path) != fh.allowedPath {
		http.NotFound(w, r)
	} else {
		fh.h.ServeHTTP(w, r)
	}
}

// loggingHandler is a wrapping handler that logs the IP of the requestor and the path of the request, as well as timing information.
type loggingHandler struct {
	h       http.Handler
	logName string
}

func NewLogging(logName string, h http.Handler) http.Handler {
	return loggingHandler{
		h:       h,
		logName: logName,
	}
}

func (lh loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lh.h.ServeHTTP(w, r)
	log.Printf("[%s] %s requested %s [took %v]", lh.logName, clientIP(r), r.URL.RequestURI(), time.Since(start))
}

func clientIP(r *http.Request) string {
	// Strip port from remote address.
	ra := r.RemoteAddr
	idx := strings.LastIndex(ra, ":")
	if idx != -1 {
		ra = ra[:idx]
	}
	return ra
}
