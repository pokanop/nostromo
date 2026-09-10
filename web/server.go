package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/task"
	"github.com/pokanop/nostromo/version"
)

//go:embed static
var staticFiles embed.FS

// csrfHeader must be sent with every mutating request. Browsers only attach
// custom headers to same-origin requests (or after a CORS preflight, which
// this server never approves) so cross-site forms cannot trigger changes.
const csrfHeader = "X-Nostromo-Web"

// Server serves the nostromo web UI and its JSON API
type Server struct {
	mu      sync.Mutex
	ver     *version.Info
	handler http.Handler
}

// NewServer returns a server for the web UI
func NewServer(ver *version.Info) *Server {
	s := &Server{ver: ver}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/meta", s.handleMeta)
	mux.HandleFunc("/api/manifests", s.handleManifests)
	mux.HandleFunc("/api/manifest", s.handleManifest)
	mux.HandleFunc("/api/command", s.handleCommand)
	mux.HandleFunc("/api/command/sub", s.handleSubstitution)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	mux.HandleFunc("/", s.handleStatic)

	s.handler = s.secure(mux)
	return s
}

// Handler returns the http handler with security middleware applied
func (s *Server) Handler() http.Handler {
	return s.handler
}

// Serve the web UI on the loopback interface until interrupted
//
// Returns an exit code suitable for os.Exit.
func Serve(port int, openBrowser bool, ver *version.Info) int {
	if _, err := task.LoadConfig(); err != nil {
		log.Error(err)
		log.Info("unable to open config file, be sure to run `nostromo init` if you haven't already")
		return -1
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Error(err)
		return -1
	}

	url := fmt.Sprintf("http://%s", ln.Addr().String())
	srv := &http.Server{
		Handler:           NewServer(ver).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	log.Highlightf("nostromo web running at %s\n", url)
	log.Regular("press ctrl+c to stop")

	if openBrowser {
		if err := launchBrowser(url); err != nil {
			log.Warningf("unable to open browser: %s\n", err)
		}
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error(err)
			return -1
		}
		log.Regular("\nnostromo web stopped")
		return 0
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(err)
			return -1
		}
		return 0
	}
}

// secure wraps the mux with same-origin and header checks
func (s *Server) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; frame-ancestors 'none'")

		// Reject DNS rebinding attempts where a public hostname resolves to loopback
		if !isLoopbackHost(r.Host) {
			writeError(w, http.StatusForbidden, "invalid host")
			return
		}

		if origin := r.Header.Get("Origin"); origin != "" && !sameOrigin(origin, r.Host) {
			writeError(w, http.StatusForbidden, "cross-origin request rejected")
			return
		}

		if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			writeError(w, http.StatusForbidden, "cross-site request rejected")
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get(csrfHeader) == "" {
				writeError(w, http.StatusForbidden, "missing "+csrfHeader+" header")
				return
			}
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}

		next.ServeHTTP(w, r)
	})
}

// handleStatic serves embedded frontend files without directory listings
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}

	b, err := fs.ReadFile(staticFiles, path.Join("static", name))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctype, ok := contentTypes[path.Ext(name)]
	if !ok {
		ctype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(b)
}

// contentTypes are fixed so platform mime databases cannot serve scripts with
// a type that nosniff would block
var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "text/javascript; charset=utf-8",
	".svg":  "image/svg+xml",
	".json": "application/json",
}

func isLoopbackHost(host string) bool {
	h := host
	if hp, _, err := net.SplitHostPort(host); err == nil {
		h = hp
	}
	h = strings.Trim(h, "[]")
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

func sameOrigin(origin, host string) bool {
	return strings.EqualFold(origin, "http://"+host)
}

func launchBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
