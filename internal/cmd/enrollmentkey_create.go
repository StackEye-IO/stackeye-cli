// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// enrollmentKeyCreateTimeout is the maximum time to wait for the API response.
const enrollmentKeyCreateTimeout = 30 * time.Second

// validEnrollmentKeyCapabilities mirrors models.ValidStationCapabilities
// (pkg/models/station.go) — checked client-side for fast feedback; the
// server is the authority.
var validEnrollmentKeyCapabilities = map[string]bool{
	"host_monitoring": true,
	"private_relay":   true,
}

// enrollmentKeyCreateFlags holds the flag values for the enrollment-key create command.
type enrollmentKeyCreateFlags struct {
	name       string
	capability []string
	mode       string
	maxUses    int
	ttlSeconds int
	unbounded  bool
}

// NewEnrollmentKeyCreateCmd creates and returns the enrollment-key create subcommand.
func NewEnrollmentKeyCreateCmd() *cobra.Command {
	flags := &enrollmentKeyCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Mint a new Station enrollment key",
		Long: `Mint a new reusable Station enrollment key.

IMPORTANT: The full enrollment key is displayed ONLY ONCE after creation.
You must save it immediately - it cannot be retrieved later.

Bounded by default: an empty invocation mints a standard-mode key with a
finite max-uses cap (100) and expiry (90 days). --unbounded is an explicit,
flagged opt-in for unlimited uses and no expiry, and cannot be combined with
--max-uses/--ttl-seconds.

Optional Flags:
  --name          Human-readable label for the key
  --capability    Capability enrolled Stations inherit (repeatable):
                  host_monitoring, private_relay
  --mode          Key class: standard (default, 100-use cap) or fleet
                  (10,000-use cap — sized for a DaemonSet)
  --max-uses      Explicit enrollment cap override (1-50000)
  --ttl-seconds   Explicit expiry override in seconds from now (up to 365 days)
  --unbounded     Explicit opt-in for unlimited uses + no expiry

Sizing: budget for nodes x re-enrollments per node over the TTL, not node
count (design docs/station/design.md §5) — a Station re-enrolls on every
pod/container replacement, so a 500-node DaemonSet restarting weekly can
consume ~7,000 uses over a 90-day TTL. Above ~1,000 nodes, set --max-uses
explicitly rather than relying on the fleet default.

Examples:
  # Mint a standard host-monitoring key
  stackeye enrollment-key create --name "VM rollout" --capability host_monitoring

  # Mint a fleet key for a large DaemonSet
  stackeye enrollment-key create --name "prod DaemonSet" --capability host_monitoring --mode fleet

  # Mint a key with an explicit cap and expiry
  stackeye enrollment-key create --name "contractor bootstrap" --max-uses 25 --ttl-seconds 604800

  # Mint an unbounded key (explicit opt-in)
  stackeye enrollment-key create --name "long-lived bootstrap" --unbounded

  # JSON output for scripting
  stackeye enrollment-key create --name "CI bootstrap" -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnrollmentKeyCreate(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.name, "name", "", "human-readable label for the key")
	cmd.Flags().StringArrayVar(&flags.capability, "capability", nil, "capability enrolled Stations inherit (repeatable): host_monitoring, private_relay")
	cmd.Flags().StringVar(&flags.mode, "mode", "", "key class: standard (default) or fleet")
	cmd.Flags().IntVar(&flags.maxUses, "max-uses", 0, "explicit enrollment cap override (1-50000)")
	cmd.Flags().IntVar(&flags.ttlSeconds, "ttl-seconds", 0, "explicit expiry override in seconds from now")
	cmd.Flags().BoolVar(&flags.unbounded, "unbounded", false, "explicit opt-in for unlimited uses + no expiry")

	return cmd
}

// runEnrollmentKeyCreate executes the enrollment-key create command logic.
func runEnrollmentKeyCreate(ctx context.Context, flags *enrollmentKeyCreateFlags) error {
	for _, capability := range flags.capability {
		if !validEnrollmentKeyCapabilities[capability] {
			return fmt.Errorf("invalid --capability %q: must be host_monitoring or private_relay", capability)
		}
	}

	if flags.mode != "" && flags.mode != string(client.EnrollmentKeyModeStandard) && flags.mode != string(client.EnrollmentKeyModeFleet) {
		return fmt.Errorf("invalid --mode %q: must be standard or fleet", flags.mode)
	}

	if flags.unbounded && (flags.maxUses != 0 || flags.ttlSeconds != 0) {
		return fmt.Errorf("--unbounded cannot be combined with --max-uses/--ttl-seconds")
	}

	req := client.CreateEnrollmentKeyRequest{
		Name:          flags.name,
		CapabilitySet: flags.capability,
		Mode:          client.EnrollmentKeyMode(flags.mode),
		Unbounded:     flags.unbounded,
	}
	if flags.maxUses > 0 {
		req.MaxUses = &flags.maxUses
	}
	if flags.ttlSeconds > 0 {
		req.TTLSeconds = &flags.ttlSeconds
	}

	if GetDryRun() {
		details := []string{
			"Name", flags.name,
			"Mode", defaultString(flags.mode, "standard"),
		}
		if len(flags.capability) > 0 {
			details = append(details, "Capabilities", fmt.Sprintf("%v", flags.capability))
		}
		if flags.unbounded {
			details = append(details, "Unbounded", "true")
		} else {
			if flags.maxUses > 0 {
				details = append(details, "Max Uses", fmt.Sprintf("%d", flags.maxUses))
			}
			if flags.ttlSeconds > 0 {
				details = append(details, "TTL Seconds", fmt.Sprintf("%d", flags.ttlSeconds))
			}
		}
		dryrun.PrintAction("create", "enrollment key", details...)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, enrollmentKeyCreateTimeout)
	defer cancel()

	result, err := client.CreateEnrollmentKey(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create enrollment key: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(result)
		}
	}

	printEnrollmentKeyMinted("ENROLLMENT KEY CREATED", result)
	return nil
}

// defaultString returns s if non-empty, else fallback.
func defaultString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// printEnrollmentKeyMinted formats and prints a newly minted/rotated
// enrollment key, showing the one-time plaintext prominently.
func printEnrollmentKeyMinted(heading string, result *client.EnrollmentKeyMintResult) {
	key := result.EnrollmentKey

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  %-60s ║\n", heading)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  ID:           %s\n", key.ID)
	if key.Name != "" {
		fmt.Printf("  Name:         %s\n", key.Name)
	}
	fmt.Printf("  Mode:         %s\n", defaultString(string(key.Mode), "standard"))
	if len(key.CapabilitySet) > 0 {
		fmt.Printf("  Capabilities: %v\n", key.CapabilitySet)
	}
	if key.MaxUses != nil {
		fmt.Printf("  Max Uses:     %d\n", *key.MaxUses)
	} else {
		fmt.Println("  Max Uses:     unlimited")
	}
	if key.ExpiresAt != nil {
		fmt.Printf("  Expires:      %s\n", *key.ExpiresAt)
	} else {
		fmt.Println("  Expires:      never")
	}
	fmt.Println()
	fmt.Println("  Plaintext Key (save this now — shown only once):")
	fmt.Println("  ┌──────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("    %s\n", result.PlaintextKey)
	fmt.Println("  └──────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ⚠ WARNING: This key will NOT be shown again — store it in a secrets manager.")
	fmt.Println()
}
