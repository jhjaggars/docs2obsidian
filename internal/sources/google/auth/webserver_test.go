package auth

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStartAuthServer(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	if server.port <= 0 {
		t.Errorf("Expected positive port number, got %d", server.port)
	}

	expectedURL := fmt.Sprintf("http://127.0.0.1:%d/callback", server.port)
	if server.getRedirectURL() != expectedURL {
		t.Errorf("Expected redirect URL %s, got %s", expectedURL, server.getRedirectURL())
	}

	if server.getState() == "" {
		t.Error("Expected non-empty state parameter")
	}

	if len(server.getState()) != 32 {
		t.Errorf("Expected state length 32, got %d", len(server.getState()))
	}
}

func TestFindAvailablePort(t *testing.T) {
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}

	if port <= 0 || port > 65535 {
		t.Errorf("Invalid port number: %d", port)
	}
}

func TestHandleCallback_Success(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	testCode := "test_auth_code_123"
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?code=%s&state=%s", server.port, testCode, server.getState())

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	select {
	case receivedCode := <-server.authCode:
		if receivedCode != testCode {
			t.Errorf("Expected auth code %s, got %s", testCode, receivedCode)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for auth code")
	}
}

func TestHandleCallback_Error(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?error=access_denied", server.port)

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	select {
	case <-server.errCh:
	case <-time.After(1 * time.Second):
		t.Error("Expected error to be sent to error channel")
	}
}

func TestHandleCallback_MissingCode(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?state=%s", server.port, server.getState())

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	select {
	case <-server.errCh:
	case <-time.After(1 * time.Second):
		t.Error("Expected error to be sent to error channel")
	}
}

func TestHandleCallback_InvalidState(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	testCode := "test_auth_code_123"
	invalidState := "invalid_state_token"
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?code=%s&state=%s", server.port, testCode, invalidState)

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	select {
	case err := <-server.errCh:
		if !strings.Contains(err.Error(), "state parameter mismatch") {
			t.Errorf("Expected state mismatch error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected error to be sent to error channel")
	}

	select {
	case <-server.authCode:
		t.Error("Should not receive auth code on invalid state")
	default:
	}
}

func TestGenerateRandomState(t *testing.T) {
	state1, err := generateRandomState()
	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	state2, err := generateRandomState()
	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	if state1 == state2 {
		t.Error("Generated states should be unique")
	}

	if len(state1) != 32 {
		t.Errorf("Expected state length 32, got %d", len(state1))
	}

	if len(state2) != 32 {
		t.Errorf("Expected state length 32, got %d", len(state2))
	}
}

func TestWaitForCode_Success(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	expectedCode := "test_code_456"

	go func() {
		time.Sleep(100 * time.Millisecond)
		server.authCode <- expectedCode
	}()

	code, err := server.waitForCode(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to wait for code: %v", err)
	}

	if code != expectedCode {
		t.Errorf("Expected code %s, got %s", expectedCode, code)
	}
}

func TestWaitForCode_Timeout(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	_, err = server.waitForCode(100 * time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestWaitForCode_Error(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	expectedError := fmt.Errorf("test error")

	go func() {
		time.Sleep(100 * time.Millisecond)
		server.errCh <- expectedError
	}()

	_, err = server.waitForCode(1 * time.Second)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestGetTokenFromWebServer_Integration(t *testing.T) {
	t.Skip("Skipping integration test to avoid network calls in unit tests")
}

func TestOpenBrowser(t *testing.T) {
	url := "https://example.com"
	
	err := openBrowser(url)

	if err != nil {
		t.Logf("Browser open failed (expected in CI/test environment): %v", err)
	}
}

func TestSuccessPageContent(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?code=test_code&state=%s", server.port, server.getState())

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "text/html" {
		t.Error("Expected Content-Type to be text/html")
	}

	bodyBytes := make([]byte, 2048)
	n, _ := resp.Body.Read(bodyBytes)
	body := string(bodyBytes[:n])

	if !strings.Contains(body, "Authorization Successful") {
		t.Error("Success page should contain 'Authorization Successful'")
	}


	if !strings.Contains(body, "pkm-sync") {
		t.Error("Success page should contain application name")
	}
}

func TestErrorPageContent(t *testing.T) {
	server, err := startAuthServer()
	if err != nil {
		t.Fatalf("Failed to start auth server: %v", err)
	}
	defer server.shutdown()

	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?error=access_denied&state=%s", server.port, server.getState())

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Failed to call callback endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "text/html" {
		t.Error("Expected Content-Type to be text/html")
	}

	bodyBytes := make([]byte, 2048)
	n, _ := resp.Body.Read(bodyBytes)
	body := string(bodyBytes[:n])

	if !strings.Contains(body, "Authorization Failed") {
		t.Error("Error page should contain 'Authorization Failed'")
	}

	if !strings.Contains(body, "access_denied") {
		t.Error("Error page should contain the error message")
	}


	if !strings.Contains(body, "pkm-sync") {
		t.Error("Error page should contain application name")
	}
}