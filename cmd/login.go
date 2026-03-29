package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/cubeapm-cli/internal/auth"
	"github.com/piyush-gambhir/cubeapm-cli/internal/client"
	"github.com/piyush-gambhir/cubeapm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Interactively configure a CubeAPM connection profile",
		Long: `Interactively configure a CubeAPM connection profile.

Prompts for:
  1. Profile name (default: "default")
  2. Server address (hostname or URL)
  3. Authentication method:
     - Email/Password -- for CubeAPM instances with authentication enabled
     - No authentication -- for unauthenticated instances (direct port access)
  4. Port configuration (query, ingest, admin)

After collecting the information, tests the connection and saves the profile
to the config file (~/.config/cubeapm-cli/config.yaml).

Examples:
  # Start interactive login
  cubeapm login

  # Login with a preset server (will still prompt for other fields)
  cubeapm login --server cubeapm.example.com`,
		Args: cobra.NoArgs,
		RunE: runLogin,
	}
}

func runLogin(cmd *cobra.Command, args []string) error {
	if cmdutil.NoInput {
		return fmt.Errorf("login requires interactive prompts; cannot run with --no-input (use 'cubeapm config set' or environment variables instead)")
	}

	reader := bufio.NewReader(os.Stdin)

	// Prompt for profile name
	fmt.Print("Profile name [default]: ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		profileName = "default"
	}

	// Prompt for server
	defaultServer := cmdutil.Resolved.Server
	if defaultServer != "" {
		fmt.Printf("CubeAPM server [%s]: ", defaultServer)
	} else {
		fmt.Print("CubeAPM server: ")
	}
	server, _ := reader.ReadString('\n')
	server = strings.TrimSpace(server)
	if server == "" {
		server = defaultServer
	}
	if server == "" {
		return fmt.Errorf("server address is required")
	}

	// Prompt for auth method
	fmt.Println("\nAuthentication:")
	fmt.Println("  1. Email/Password")
	fmt.Println("  2. No authentication")
	fmt.Print("Choose [1]: ")
	authChoice, _ := reader.ReadString('\n')
	authChoice = strings.TrimSpace(authChoice)
	if authChoice == "" {
		authChoice = "1"
	}

	var (
		authMethod    string
		email         string
		password      string
		sessionCookie string
		sessionExpiry string
	)

	switch authChoice {
	case "1":
		authMethod = "kratos"

		// Prompt for email
		fmt.Print("Email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email == "" {
			return fmt.Errorf("email is required for email/password authentication")
		}

		// Prompt for password (hidden input)
		fmt.Print("Password: ")
		if term.IsTerminal(int(os.Stdin.Fd())) {
			pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println() // newline after hidden input
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			password = string(pwBytes)
		} else {
			// Fallback for piped input
			password, _ = reader.ReadString('\n')
			password = strings.TrimSpace(password)
		}
		if password == "" {
			return fmt.Errorf("password is required for email/password authentication")
		}

		// Test Kratos login immediately
		if !cmdutil.Quiet {
			fmt.Print("\nAuthenticating... ")
		}

		// Build the server base URL for Kratos
		serverURL := server
		if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
			serverURL = "https://" + serverURL
		}

		session, err := auth.KratosLogin(serverURL, email, password, flagVerbose)
		if err != nil {
			if !cmdutil.Quiet {
				fmt.Println("FAILED")
			}
			return fmt.Errorf("authentication failed: %w", err)
		}
		sessionCookie = session.Cookie
		sessionExpiry = session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")

		if !cmdutil.Quiet {
			fmt.Printf("OK (session expires %s)\n", session.ExpiresAt.Format("2006-01-02 15:04 MST"))
		}

	case "2":
		authMethod = "none"

	default:
		return fmt.Errorf("invalid choice %q: enter 1 or 2", authChoice)
	}

	// Prompt for query port
	fmt.Printf("Query port [%d]: ", config.DefaultQueryPort)
	queryPortStr, _ := reader.ReadString('\n')
	queryPortStr = strings.TrimSpace(queryPortStr)
	queryPort := config.DefaultQueryPort
	if queryPortStr != "" {
		fmt.Sscanf(queryPortStr, "%d", &queryPort)
	}

	// Prompt for ingest port
	fmt.Printf("Ingest port [%d]: ", config.DefaultIngestPort)
	ingestPortStr, _ := reader.ReadString('\n')
	ingestPortStr = strings.TrimSpace(ingestPortStr)
	ingestPort := config.DefaultIngestPort
	if ingestPortStr != "" {
		fmt.Sscanf(ingestPortStr, "%d", &ingestPort)
	}

	// Prompt for admin port
	fmt.Printf("Admin port [%d]: ", config.DefaultAdminPort)
	adminPortStr, _ := reader.ReadString('\n')
	adminPortStr = strings.TrimSpace(adminPortStr)
	adminPort := config.DefaultAdminPort
	if adminPortStr != "" {
		fmt.Sscanf(adminPortStr, "%d", &adminPort)
	}

	// Test connection with the configured auth
	if !cmdutil.Quiet && authMethod != "kratos" {
		fmt.Print("\nTesting connection... ")
	}

	testCfg := config.ResolvedConfig{
		Server:        server,
		QueryPort:     queryPort,
		IngestPort:    ingestPort,
		AdminPort:     adminPort,
		AuthMethod:    authMethod,
		Email:         email,
		Password:      password,
		SessionCookie: sessionCookie,
	}
	testClient, err := client.NewClient(testCfg)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	services, err := testClient.GetServices()
	if err != nil {
		if !cmdutil.Quiet && authMethod != "kratos" {
			fmt.Println("FAILED")
		}
		return fmt.Errorf("connection test failed: %w", err)
	}
	if !cmdutil.Quiet {
		if authMethod != "kratos" {
			fmt.Printf("OK (%d services found)\n", len(services))
		} else {
			fmt.Printf("Connection verified (%d services found)\n", len(services))
		}
	}

	// Save profile
	profile := config.Profile{
		Server:        server,
		QueryPort:     queryPort,
		IngestPort:    ingestPort,
		AdminPort:     adminPort,
		AuthMethod:    authMethod,
		Email:         email,
		Password:      password,
		SessionCookie: sessionCookie,
		SessionExpiry: sessionExpiry,
	}

	cmdutil.AppConfig.SetProfile(profileName, profile)
	if cmdutil.AppConfig.CurrentProfile == "" {
		cmdutil.AppConfig.CurrentProfile = profileName
	}

	if err := config.Save(cmdutil.AppConfig); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if !cmdutil.Quiet {
		fmt.Printf("\nProfile %q saved", profileName)
		if cmdutil.AppConfig.CurrentProfile == profileName {
			fmt.Print(" (active)")
		}
		fmt.Println()
		fmt.Printf("Config written to %s\n", config.ConfigPath())
	}

	return nil
}
