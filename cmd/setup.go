package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"docs2obsidian/internal/auth"
	"docs2obsidian/internal/calendar"
	"docs2obsidian/internal/config"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify authentication configuration",
	Long:  "Validates OAuth 2.0 credentials and tests API access to ensure everything is configured correctly.",
	RunE:  runSetupCommand,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetupCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("Validating OAuth 2.0 authentication configuration...")
	fmt.Println()

	fmt.Println("1. Checking for credentials.json file...")
	
	credentialsPath, err := config.FindCredentialsFile()
	if err != nil {
		fmt.Println("   [FAIL] No credentials.json file found")
		fmt.Println()
		
		defaultPath, _ := config.GetCredentialsPath()
		fmt.Printf("Searched in:\n")
		fmt.Printf("  - %s (default config directory)\n", defaultPath)
		fmt.Printf("  - ./credentials.json (current directory)\n")
		fmt.Println()
		
		fmt.Println("[FAIL] OAuth 2.0 credentials not configured!")
		fmt.Println()
		fmt.Println("To set up authentication:")
		fmt.Println("1. Go to the Google Cloud Console (https://console.cloud.google.com/)")
		fmt.Println("2. Create a new project or select an existing one")
		fmt.Println("3. Enable the Google Calendar API and Google Drive API")
		fmt.Println("4. Configure the OAuth consent screen")
		fmt.Println("5. Create OAuth 2.0 Client ID credentials for a 'Desktop application'")
		fmt.Printf("6. Download the credentials and save as 'credentials.json' in: %s\n", defaultPath)
		fmt.Println()
		fmt.Println("Alternatively, use a custom path with:")
		fmt.Println("  docs2obsidian --credentials /path/to/credentials.json setup")
		
		return fmt.Errorf("credentials.json file not found")
	}
	fmt.Printf("   [OK] Found credentials.json at: %s\n", credentialsPath)
	
	fmt.Println()
	fmt.Println("2. Testing OAuth 2.0 flow and API access...")
	
	client, err := auth.GetClient()
	if err != nil {
		fmt.Printf("   [FAIL] Failed to authenticate: %v\n", err)
		fmt.Println()
		fmt.Println("This usually means:")
		fmt.Println("- Invalid credentials.json file format")
		fmt.Println("- OAuth consent screen not properly configured")
		fmt.Println("- User denied access during OAuth flow")
		fmt.Println()
		fmt.Println("Check your credentials.json file and try again.")
		return fmt.Errorf("OAuth 2.0 authentication failed: %w", err)
	}
	
	calendarService, err := calendar.NewService(client)
	if err != nil {
		fmt.Printf("   [FAIL] Failed to create calendar service: %v\n", err)
		return fmt.Errorf("calendar service creation failed: %w", err)
	}
	
	events, err := calendarService.GetUpcomingEvents(1)
	if err != nil {
		fmt.Printf("   [FAIL] Failed to access calendar: %v\n", err)
		fmt.Println()
		fmt.Println("This usually means:")
		fmt.Println("- Calendar API is not enabled in your Google Cloud project")
		fmt.Println("- OAuth consent screen doesn't include Calendar scope")
		fmt.Println("- Your Google account doesn't have calendar access")
		fmt.Println()
		fmt.Println("Verify that the Calendar API is enabled in your Google Cloud project.")
		
		return fmt.Errorf("calendar API access failed: %w", err)
	}
	
	fmt.Printf("   [OK] Successfully accessed calendar (found %d upcoming events)\n", len(events))
	fmt.Println()
	fmt.Println("All authentication checks passed!")
	fmt.Println("You can now run 'docs2obsidian calendar' to list your events.")
	
	return nil
}