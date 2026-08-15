// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewEnrollmentKeyCmd creates and returns the enrollment-key parent command.
// This command provides management of Station enrollment keys — reusable
// bootstrap credentials that mint new Stations (design docs/station/design.md §5).
func NewEnrollmentKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enrollment-key",
		Short: "Manage Station enrollment keys",
		Long: `Manage reusable Station enrollment keys.

An enrollment key (se_ek_...) is a bootstrap-only credential that mints new
Stations; it is never a Station's ongoing identity credential. Keys are
bounded by default (finite max uses + expiry) — unlimited is an explicit
opt-in.

Key Concepts:
  - Keys are scoped to an organization
  - "mode" selects the bounded default cap: standard (100 uses, default) or
    fleet (10,000 uses — sized for a DaemonSet that re-enrolls on every pod
    restart)
  - capability_set is inherited by every Station enrolled with the key
  - Revoking a key stops NEW enrollments; already-enrolled Stations are untouched
  - Keys are shown only once at creation/rotation — store them securely

Sizing (design docs/station/design.md §5): the budget a key must cover is
nodes x expected re-enrollments per node over the TTL, not node count — a
Station re-enrolls on every pod/container replacement. Above ~1,000 nodes,
set --max-uses explicitly rather than relying on the fleet default.

Available Commands:
  list        List all enrollment keys for the current organization
  create      Mint a new enrollment key
  revoke      Revoke an existing enrollment key
  rotate      Rotate an enrollment key (mint fresh + revoke old, atomically)

Examples:
  # List all enrollment keys
  stackeye enrollment-key list

  # Mint a standard key for host monitoring
  stackeye enrollment-key create --name "VM rollout" --capability host_monitoring

  # Mint a fleet key for a DaemonSet
  stackeye enrollment-key create --name "prod DaemonSet" --capability host_monitoring --mode fleet

  # Mint an unbounded key (explicit opt-in — no expiry, no use cap)
  stackeye enrollment-key create --name "long-lived bootstrap" --unbounded

  # Revoke a key by ID
  stackeye enrollment-key revoke <key-id>

  # Rotate a key (mints a fresh key inheriting the old one's scoping)
  stackeye enrollment-key rotate <key-id>

For more information about a specific command:
  stackeye enrollment-key [command] --help`,
		Aliases: []string{"enrollmentkey", "enrollment-keys", "ek"},
	}

	// Register subcommands
	cmd.AddCommand(NewEnrollmentKeyListCmd())
	cmd.AddCommand(NewEnrollmentKeyCreateCmd())
	cmd.AddCommand(NewEnrollmentKeyRevokeCmd())
	cmd.AddCommand(NewEnrollmentKeyRotateCmd())

	return cmd
}
