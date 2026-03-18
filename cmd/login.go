package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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
  2. Server address (hostname or IP)
  3. API token (optional, for authenticated instances)
  4. Query port (default: 3140)
  5. Ingest port (default: 3130)
  6. Admin port (default: 3199)

After collecting the information, tests the connection by querying the
services endpoint. If successful, saves the profile to the config file
(~/.config/cubeapm/config.yaml).

You can have multiple profiles for different CubeAPM instances and switch
between them with 'cubeapm config profiles use <name>'.

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

	// Prompt for token
	fmt.Print("API token (leave empty for no auth): ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

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

	// Test connection
	if !cmdutil.Quiet {
		fmt.Print("\nTesting connection... ")
	}
	testCfg := config.ResolvedConfig{
		Server:     server,
		QueryPort:  queryPort,
		IngestPort: ingestPort,
		AdminPort:  adminPort,
		Token:      token,
	}
	testClient, err := client.NewClient(testCfg)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	services, err := testClient.GetServices()
	if err != nil {
		if !cmdutil.Quiet {
			fmt.Println("FAILED")
		}
		return fmt.Errorf("connection test failed: %w", err)
	}
	if !cmdutil.Quiet {
		fmt.Printf("OK (%d services found)\n", len(services))
	}

	// Save profile
	profile := config.Profile{
		Server:     server,
		QueryPort:  queryPort,
		IngestPort: ingestPort,
		AdminPort:  adminPort,
		Token:      token,
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
