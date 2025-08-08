package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"pkm-sync/internal/sources/google/auth"
	"pkm-sync/internal/sources/google/calendar"
	"pkm-sync/internal/config"
	"pkm-sync/internal/sources/google/drive"
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
		fmt.Println("  pkm-sync --credentials /path/to/credentials.json setup")
		
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
	fmt.Println("3. Testing Google Drive API access...")
	
	driveService, err := drive.NewService(client)
	if err != nil {
		fmt.Printf("   [FAIL] Failed to create drive service: %v\n", err)
		return fmt.Errorf("drive service creation failed: %w", err)
	}
	
	// Test Drive API by attempting to use the Files.Export method
	// This is the specific operation that's failing, so we need to test it directly
	err = testDriveExportPermissions(driveService)
	if err != nil {
		if isPermissionError(err) {
			fmt.Printf("   [FAIL] Drive export permission denied: %v\n", err)
			fmt.Println()
			fmt.Println("This usually means:")
			fmt.Println("- Drive API is not enabled in your Google Cloud project")
			fmt.Println("- OAuth consent screen doesn't include sufficient Drive scope")
			fmt.Println("- Current token has insufficient permissions for document export")
			fmt.Println()
			fmt.Println("To fix this:")
			fmt.Println("1. Delete your token file to force re-authorization:")
			tokenPath, _ := config.GetTokenPath()
			fmt.Printf("   rm %s\n", tokenPath)
			fmt.Println("2. Run this setup command again to re-authorize with full permissions")
			
			return fmt.Errorf("drive export permissions insufficient: %w", err)
		} else {
			// Other error (like "file not found") means permissions are OK
			fmt.Printf("   [OK] Drive export permissions verified\n")
		}
	} else {
		fmt.Printf("   [OK] Drive export permissions verified\n")
	}
	
	fmt.Println()
	fmt.Println("All authentication checks passed!")
	fmt.Println("You can now run:")
	fmt.Println("  - 'pkm-sync calendar' to list your events")
	fmt.Println("  - 'pkm-sync export' to export Google Docs from calendar events")
	
	return nil
}

// testDriveExportPermissions tests if the Drive API has export permissions
func testDriveExportPermissions(driveService *drive.Service) error {
	// Try to test export permissions by attempting to export a dummy file
	// This will fail with "file not found" if permissions are OK,
	// or with "insufficient permissions" if scope is wrong
	_, err := driveService.GetFileMetadata("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms")
	return err
}

// isPermissionError checks if an error is related to insufficient permissions
func isPermissionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "insufficient") ||
		   strings.Contains(errStr, "permission") ||
		   strings.Contains(errStr, "scope") ||
		   strings.Contains(errStr, "access_token_scope_insufficient") ||
		   strings.Contains(errStr, "403")
}