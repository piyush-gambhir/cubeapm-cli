package cmd

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAgentSafetyCommandManifest makes every runnable command part of a
// reviewed safety contract. Adding, removing, or reclassifying a command
// requires an intentional manifest-hash update.
func TestAgentSafetyCommandManifest(t *testing.T) {
	dangerousVerbs := map[string]bool{
		"add": true, "apply": true, "approve": true, "archive": true,
		"assign": true, "attach": true, "build": true, "cancel": true,
		"close": true, "comment": true, "create": true, "delete": true,
		"disable": true, "enable": true, "execute": true, "import": true,
		"link": true, "move": true, "open": true, "password": true,
		"permissions": true, "reload": true, "remove": true, "renew": true,
		"restart": true, "restore": true, "run": true, "set": true,
		"stop": true, "transition": true, "trigger": true, "update": true,
		"worklog": true,
	}

	var entries []string
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		for _, command := range parent.Commands() {
			if command.Runnable() {
				mutates := command.Annotations != nil && command.Annotations["mutates"] == "true"
				interactive := command.Annotations != nil && command.Annotations["interactive"] == "true"
				leaf := strings.Fields(command.Use)
				if len(leaf) > 0 && dangerousVerbs[leaf[0]] && !mutates {
					t.Errorf("%q looks mutating but is not annotated with mutates=true", command.CommandPath())
				}
				entries = append(entries, fmt.Sprintf("%s|mutates=%t|interactive=%t", command.CommandPath(), mutates, interactive))
			}
			walk(command)
		}
	}
	walk(rootCmd)
	sort.Strings(entries)
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(entries, "\n"))))
	const expectedDigest = "ebee7072bd802901e86964b6190539fcdbd6708ce6b6445e3535bdd396ff63a6"
	if digest != expectedDigest {
		t.Fatalf("agent-safety command manifest changed: got %s; review command mutation/interaction annotations, then update expectedDigest", digest)
	}
}
