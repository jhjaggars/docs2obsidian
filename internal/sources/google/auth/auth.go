package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	drive "google.golang.org/api/drive/v3"

	"pkm-sync/internal/config"
)

func GetClient() (*http.Client, error) {
	config, err := getOAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth config: %w", err)
	}

	token, err := getToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %w", err)
	}

	return config.Client(context.Background(), token), nil
}

func getOAuthConfig() (*oauth2.Config, error) {
	credentialsPath, err := config.FindCredentialsFile()
	if err != nil {
		return nil, fmt.Errorf("unable to find credentials file: %w", err)
	}

	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	oauthConfig, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	return oauthConfig, nil
}

func getToken(oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	token, err := tokenFromFile()
	if err != nil {
		// No existing token, get new one
		token, err = getTokenFromWeb(oauthConfig)
		if err != nil {
			return nil, err
		}
		saveToken(token)

		return token, nil
	}

	// Check if we have a valid access token or refresh token
	// The OAuth2 client will automatically refresh if needed
	if token.AccessToken == "" && token.RefreshToken == "" {
		// Token is completely invalid, need to re-authorize
		fmt.Println("Token is invalid. Re-authorization required.")

		token, err = getTokenFromWeb(oauthConfig)
		if err != nil {
			return nil, err
		}
		saveToken(token)
	}

	return token, nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	fmt.Println("Starting OAuth 2.0 authorization flow...")

	token, err := getTokenFromWebServer(config)
	if err != nil {
		fmt.Printf("Web server authorization failed: %v\n", err)
		fmt.Println("Falling back to manual authorization...")
		fmt.Println()

		return getTokenFromWebManual(config)
	}

	return token, nil
}

func getTokenFromWebManual(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("To authorize this application, please visit this URL in your browser:")
	fmt.Printf("%s\n\n", authURL)

	fmt.Println("After authorization, you will be redirected to a URL that looks like:")
	fmt.Println("http://localhost/callback?code=AUTHORIZATION_CODE&scope=...")
	fmt.Println()
	fmt.Println("Please copy and paste either:")
	fmt.Println("1. The full redirect URL, OR")
	fmt.Println("2. Just the authorization code from the 'code=' parameter")
	fmt.Print("Paste here: ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return nil, fmt.Errorf("unable to read input: %w", err)
	}

	authCode := extractAuthCode(input)
	if authCode == "" {
		return nil, fmt.Errorf("could not extract authorization code from input")
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
}

func extractAuthCode(input string) string {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "code=") {
		parts := strings.Split(input, "code=")
		if len(parts) > 1 {
			codePart := parts[1]
			if strings.Contains(codePart, "&") {
				codePart = strings.Split(codePart, "&")[0]
			}

			return codePart
		}
	}

	if !strings.Contains(input, "://") && !strings.Contains(input, "=") {
		return input
	}

	return ""
}

func tokenFromFile() (*oauth2.Token, error) {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)

	return token, err
}

func saveToken(token *oauth2.Token) {
	tokenPath, err := config.GetTokenPath()
	if err != nil {
		log.Fatalf("Unable to get token path: %v", err)
	}

	fmt.Printf("Saving credential file to: %s\n", tokenPath)
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
