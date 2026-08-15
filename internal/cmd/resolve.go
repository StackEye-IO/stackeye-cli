// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
)

// probeResolveTimeout is the maximum time to wait for probe resolution via API search.
const probeResolveTimeout = 10 * time.Second

// ResolveProbeID resolves a probe identifier to a UUID.
// If the input is a valid UUID, it returns it immediately without API calls.
// If the input is not a UUID, it searches for a probe by name and returns its UUID.
// Returns an error if the probe is not found or if multiple probes match the name.
func ResolveProbeID(ctx context.Context, c *client.Client, idOrName string) (uuid.UUID, error) {
	// Try UUID parse first (fast path - no API call needed)
	if probeID, err := uuid.Parse(idOrName); err == nil {
		return probeID, nil
	}

	// Not a valid UUID, search by name
	probe, err := resolveProbeByName(ctx, c, idOrName)
	if err != nil {
		return uuid.Nil, err
	}

	return probe.ID, nil
}

// ResolveProbeIDs resolves multiple probe identifiers to UUIDs.
// Each identifier can be either a UUID or a probe name.
// Returns an error if any probe cannot be resolved.
func ResolveProbeIDs(ctx context.Context, c *client.Client, idOrNames []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(idOrNames))

	for _, idOrName := range idOrNames {
		probeID, err := ResolveProbeID(ctx, c, idOrName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %q: %w", idOrName, err)
		}
		result = append(result, probeID)
	}

	return result, nil
}

// resolveProbeByName searches for a probe by name and returns it.
// Returns an error if no probe matches or if multiple probes match (ambiguous).
func resolveProbeByName(ctx context.Context, c *client.Client, name string) (*client.Probe, error) {
	// Use a shorter timeout for search operations
	reqCtx, cancel := context.WithTimeout(ctx, probeResolveTimeout)
	defer cancel()

	// Search for probes matching the name
	opts := &client.ListProbesOptions{
		Search: name,
		Limit:  100, // Get enough results to detect ambiguity
	}

	response, err := client.ListProbes(reqCtx, c, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search probes: %w", err)
	}

	if len(response.Probes) == 0 {
		return nil, fmt.Errorf("probe %q not found", name)
	}

	// Look for exact match first (case-insensitive)
	var exactMatches []client.Probe
	for _, p := range response.Probes {
		if strings.EqualFold(p.Name, name) {
			exactMatches = append(exactMatches, p)
		}
	}

	// If we have exactly one exact match, use it
	if len(exactMatches) == 1 {
		return &exactMatches[0], nil
	}

	// If we have multiple exact matches, it's ambiguous
	if len(exactMatches) > 1 {
		return nil, formatAmbiguousError(name, exactMatches)
	}

	// No exact matches - if we have exactly one substring match, use it
	if len(response.Probes) == 1 {
		return &response.Probes[0], nil
	}

	// Multiple substring matches - ambiguous
	return nil, formatAmbiguousError(name, response.Probes)
}

// formatAmbiguousError creates a user-friendly error message for ambiguous probe names.
func formatAmbiguousError(name string, matches []client.Probe) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "ambiguous probe name %q: found %d matches\n", name, len(matches))

	// Show up to 5 matches with their IDs
	limit := 5
	if len(matches) < limit {
		limit = len(matches)
	}

	for i := 0; i < limit; i++ {
		p := matches[i]
		fmt.Fprintf(&sb, "  - %s (%s)\n", p.Name, p.ID)
	}

	if len(matches) > 5 {
		fmt.Fprintf(&sb, "  ... and %d more\n", len(matches)-5)
	}

	sb.WriteString("Use the full UUID to specify the exact probe")

	return fmt.Errorf("%s", sb.String())
}

// deviceResolveTimeout is the maximum time to wait for device resolution via
// the API. Task stackeye-5859.
const deviceResolveTimeout = 10 * time.Second

// ResolveDeviceID resolves a device identifier to a UUID.
// If the input is a valid UUID, it returns it immediately without API calls.
// If the input is not a UUID, it looks up a device by exact (case-insensitive)
// name match and returns its UUID. Unlike ResolveProbeID, the devices API has
// no server-side search filter, so this fetches the organization's full
// device list and matches client-side — deliberately exact-match only (no
// substring fallback) to keep results predictable across large fleets.
// Task stackeye-5859.
func ResolveDeviceID(ctx context.Context, c *client.Client, idOrName string) (uuid.UUID, error) {
	// Try UUID parse first (fast path - no API call needed)
	if deviceID, err := uuid.Parse(idOrName); err == nil {
		return deviceID, nil
	}

	device, err := resolveDeviceByName(ctx, c, idOrName)
	if err != nil {
		return uuid.Nil, err
	}

	return device.ID, nil
}

// resolveDeviceByName looks up a device by exact (case-insensitive) name.
// Returns an error if no device matches or if multiple devices share the name.
func resolveDeviceByName(ctx context.Context, c *client.Client, name string) (*client.Device, error) {
	reqCtx, cancel := context.WithTimeout(ctx, deviceResolveTimeout)
	defer cancel()

	devices, err := client.ListDevices(reqCtx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	var exactMatches []client.Device
	for _, d := range devices {
		if strings.EqualFold(d.Name, name) {
			exactMatches = append(exactMatches, d)
		}
	}

	if len(exactMatches) == 0 {
		return nil, fmt.Errorf("device %q not found", name)
	}
	if len(exactMatches) == 1 {
		return &exactMatches[0], nil
	}

	return nil, formatAmbiguousDeviceError(name, exactMatches)
}

// formatAmbiguousDeviceError creates a user-friendly error message for
// ambiguous device names.
func formatAmbiguousDeviceError(name string, matches []client.Device) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "ambiguous device name %q: found %d matches\n", name, len(matches))

	limit := 5
	if len(matches) < limit {
		limit = len(matches)
	}

	for i := 0; i < limit; i++ {
		d := matches[i]
		fmt.Fprintf(&sb, "  - %s (%s)\n", d.Name, d.ID)
	}

	if len(matches) > 5 {
		fmt.Fprintf(&sb, "  ... and %d more\n", len(matches)-5)
	}

	sb.WriteString("Use the full UUID to specify the exact device")

	return fmt.Errorf("%s", sb.String())
}
