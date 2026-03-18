package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/cubeapm-cli/internal/config"
)

// checkReadOnly mirrors the enforcement logic in PersistentPreRunE.
// It returns an error if the command is blocked in read-only mode.
func checkReadOnly(cmd *cobra.Command, resolved config.ResolvedConfig, flagChanged bool, flagValue bool) error {
	effectiveReadOnly := resolved.ReadOnly
	if flagChanged {
		effectiveReadOnly = flagValue
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode.\nTo disable, use --read-only=false or remove read_only from your config profile.", cmd.CommandPath())
	}
	return nil
}

func TestCheckReadOnly_WriteCmdBlocked(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete-things",
		Annotations: map[string]string{"mutates": "true"},
	}

	resolved := config.ResolvedConfig{ReadOnly: true}

	err := checkReadOnly(cmd, resolved, false, false)
	if err == nil {
		t.Fatal("expected error for write command in read-only mode, got nil")
	}
}

func TestCheckReadOnly_WriteCmdAllowed(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete-things",
		Annotations: map[string]string{"mutates": "true"},
	}

	// read-only is false in resolved config
	resolved := config.ResolvedConfig{ReadOnly: false}

	err := checkReadOnly(cmd, resolved, false, false)
	if err != nil {
		t.Fatalf("expected no error for write command with read-only=false, got: %v", err)
	}
}

func TestCheckReadOnly_ReadCmdAllowed(t *testing.T) {
	cmd := &cobra.Command{
		Use: "list-things",
		// No "mutates" annotation — this is a read command.
	}

	resolved := config.ResolvedConfig{ReadOnly: true}

	err := checkReadOnly(cmd, resolved, false, false)
	if err != nil {
		t.Fatalf("expected no error for read command in read-only mode, got: %v", err)
	}
}

func TestCheckReadOnly_FlagOverridesConfig(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete-things",
		Annotations: map[string]string{"mutates": "true"},
	}

	// Config says read-only=false, but flag overrides to true
	resolved := config.ResolvedConfig{ReadOnly: false}

	err := checkReadOnly(cmd, resolved, true, true)
	if err == nil {
		t.Fatal("expected error when --read-only flag overrides config to true")
	}

	// Config says read-only=true, but flag overrides to false
	resolved.ReadOnly = true

	err = checkReadOnly(cmd, resolved, true, false)
	if err != nil {
		t.Fatalf("expected no error when --read-only=false flag overrides config, got: %v", err)
	}
}
