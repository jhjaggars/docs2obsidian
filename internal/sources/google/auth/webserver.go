package auth

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	successTemplate *template.Template
	errorTemplate   *template.Template
)

func init() {
	var err error
	successTemplate, err = template.ParseFS(templateFS, "templates/success.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse success template: %v", err))
	}

	errorTemplate, err = template.ParseFS(templateFS, "templates/error.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse error template: %v", err))
	}
}

type authServer struct {
	server   *http.Server
	port     int
	authCode chan string
	errCh    chan error
	state    string
}

func startAuthServer() (*authServer, error) {
	port, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	state, err := generateRandomState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state parameter: %w", err)
	}

	as := &authServer{
		port:     port,
		authCode: make(chan string, 1),
		errCh:    make(chan error, 1),
		state:    state,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", as.handleCallback)

	as.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	go func() {
		if err := as.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			as.errCh <- err
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return as, nil
}

func (as *authServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		as.serveErrorPage(w, errorParam)
		as.errCh <- fmt.Errorf("OAuth error: %s", errorParam)

		return
	}

	if state != as.state {
		as.serveErrorPage(w, "invalid state parameter - possible CSRF attack")
		as.errCh <- fmt.Errorf("state parameter mismatch: expected %s, got %s", as.state, state)

		return
	}

	if code == "" {
		as.serveErrorPage(w, "missing authorization code")
		as.errCh <- fmt.Errorf("authorization code not found in callback")

		return
	}

	as.serveSuccessPage(w)
	as.authCode <- code
}

func (as *authServer) serveSuccessPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	if err := successTemplate.Execute(w, nil); err != nil {
		as.errCh <- fmt.Errorf("failed to execute success template: %w", err)
	}
}

func (as *authServer) serveErrorPage(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)

	if err := errorTemplate.Execute(w, errorMsg); err != nil {
		as.errCh <- fmt.Errorf("failed to execute error template: %w", err)
	}
}

func (as *authServer) waitForCode(timeout time.Duration) (string, error) {
	select {
	case code := <-as.authCode:
		return code, nil
	case err := <-as.errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for authorization")
	}
}

func (as *authServer) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return as.server.Shutdown(ctx)
}

func (as *authServer) getRedirectURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", as.port)
}

func (as *authServer) getState() string {
	return as.state
}

func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	return addr.Port, nil
}

func getTokenFromWebServer(config *oauth2.Config) (*oauth2.Token, error) {
	server, err := startAuthServer()
	if err != nil {
		return nil, fmt.Errorf("failed to start auth server: %w", err)
	}
	defer server.shutdown()

	config.RedirectURL = server.getRedirectURL()
	authURL := config.AuthCodeURL(server.getState(), oauth2.AccessTypeOffline)

	fmt.Println("Opening authorization URL in your browser...")
	fmt.Printf("If your browser doesn't open automatically, visit: %s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
		fmt.Println("Please open the URL manually in your browser.")
	}

	authCode, err := server.waitForCode(5 * time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	fmt.Println("Authorization successful!")

	return token, nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case strings.Contains(strings.ToLower(os.Getenv("OSTYPE")), "darwin") || strings.Contains(strings.ToLower(os.Getenv("PATH")), "darwin"):
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
