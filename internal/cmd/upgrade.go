// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/update"
	"github.com/StackEye-IO/stackeye-cli/internal/version"
	sdkupdate "github.com/StackEye-IO/stackeye-go-sdk/update"
	"github.com/spf13/cobra"
)

// upgradeFlags holds the command flags for the upgrade command.
type upgradeFlags struct {
	version string
	force   bool
	dryRun  bool
}

// NewUpgradeCmd creates and returns the upgrade command.
func NewUpgradeCmd() *cobra.Command {
	flags := &upgradeFlags{}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade to the latest version of StackEye CLI",
		Long: `Upgrade to the latest version of StackEye CLI by downloading and
installing the latest release from GitHub.

The command will:
  - Check for available updates
  - Download the appropriate binary for your platform
  - Verify the SHA256 checksum
  - Replace the current binary

By default, the command only upgrades if a newer version is available.
Use --force to reinstall the current version.

Examples:
  # Upgrade to the latest version
  stackeye upgrade

  # Install a specific version
  stackeye upgrade --version v1.2.3

  # Force reinstall even if already on latest
  stackeye upgrade --force

  # Preview what would happen without making changes
  stackeye upgrade --dry-run`,
		// Override PersistentPreRunE to skip config loading requirement.
		// The upgrade command should work without valid configuration
		// so users can upgrade even if their config is corrupted.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.version, "version", "", "install a specific version (e.g., v1.2.3)")
	cmd.Flags().BoolVar(&flags.force, "force", false, "reinstall even if already on latest version")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "show what would be done without making changes")

	return cmd
}

// runUpgrade executes the upgrade operation.
func runUpgrade(ctx context.Context, flags *upgradeFlags) error {
	currentVersion := version.Version

	// Check if this is a dev build
	if currentVersion == "dev" && !flags.force {
		fmt.Println("Cannot upgrade development builds.")
		fmt.Println("Use --force to override or install from GitHub releases.")
		return nil
	}

	// Create updater
	updater := sdkupdate.NewUpdater(update.GitHubRepo, currentVersion)

	// Determine which release to fetch
	var release *sdkupdate.ReleaseInfo
	var err error

	if flags.version != "" {
		// Fetch specific version
		fmt.Printf("Fetching release %s...\n", flags.version)
		release, err = getSpecificRelease(ctx, updater, flags.version)
		if err != nil {
			return fmt.Errorf("failed to get release %s: %w", flags.version, err)
		}
		if release == nil {
			return fmt.Errorf("release %s not found", flags.version)
		}
	} else {
		// Check for latest release
		fmt.Println("Checking for updates...")
		if flags.force {
			release, err = updater.GetLatestRelease(ctx)
		} else {
			release, err = updater.CheckForUpdate(ctx)
		}
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
	}

	// Check if update is available (unless forcing or specific version)
	if release == nil && flags.version == "" {
		fmt.Printf("Already up to date (version %s).\n", currentVersion)
		return nil
	}

	// Find appropriate asset for this platform
	asset := release.FindAssetForPlatform()
	if asset == nil {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Display upgrade information
	fmt.Println()
	fmt.Printf("  Current version: %s\n", currentVersion)
	fmt.Printf("  New version:     %s\n", release.Version)
	fmt.Printf("  Asset:           %s\n", asset.Name)
	fmt.Printf("  Size:            %s\n", formatSize(asset.Size))
	fmt.Println()

	// Dry run stops here
	if flags.dryRun {
		fmt.Println("Would download and install this version.")
		fmt.Println("Run without --dry-run to proceed.")
		return nil
	}

	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Check if we have write permission to the executable
	if err := checkWritePermission(execPath); err != nil {
		return fmt.Errorf("cannot update %s: %w", execPath, err)
	}

	// Create a temporary directory for the download
	tmpDir, err := os.MkdirTemp("", "stackeye-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the release with progress reporting
	fmt.Printf("Downloading %s...\n", asset.Name)

	progressBar := newByteProgressBar(asset.Size)
	progressBar.Start()

	downloadCfg := sdkupdate.DownloadConfig{
		VerifyChecksum: true,
		ProgressFunc: func(downloaded, total int64) {
			progressBar.Update(downloaded)
		},
	}

	result, err := sdkupdate.DownloadRelease(ctx, asset, tmpDir, downloadCfg)
	progressBar.Finish()

	if err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}

	// Report checksum verification
	if result.ChecksumVerified {
		fmt.Println("Checksum verified.")
	}

	// Replace the executable
	fmt.Println("Installing...")
	if err := replaceExecutable(execPath, result.FilePath); err != nil {
		return fmt.Errorf("failed to install new version: %w", err)
	}

	fmt.Printf("\nSuccessfully upgraded to version %s!\n", release.Version)

	if release.HTMLURL != "" {
		fmt.Printf("Release notes: %s\n", release.HTMLURL)
	}

	return nil
}

// getSpecificRelease fetches a specific release by tag name.
func getSpecificRelease(ctx context.Context, updater *sdkupdate.Updater, tag string) (*sdkupdate.ReleaseInfo, error) {
	// Validate tag is not empty
	if tag == "" {
		return nil, fmt.Errorf("version tag cannot be empty")
	}

	// Normalize the tag (ensure it starts with 'v')
	if tag[0] != 'v' {
		tag = "v" + tag
	}

	// For specific versions, we need to query the releases API directly
	// Create a new updater with the specific version to leverage existing code
	specificUpdater := sdkupdate.NewUpdater(updater.Repository(), tag)
	return specificUpdater.GetLatestRelease(ctx)
}

// checkWritePermission verifies we can write to the executable location.
func checkWritePermission(path string) error {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check if we have write permission by trying to open for writing
	f, err := os.OpenFile(path, os.O_WRONLY, info.Mode())
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied (try running with sudo)")
		}
		return err
	}
	f.Close()

	return nil
}

// replaceExecutable atomically replaces the current executable with the new one.
func replaceExecutable(oldPath, newPath string) error {
	// Get the permissions from the old executable
	info, err := os.Stat(oldPath)
	if err != nil {
		return fmt.Errorf("failed to stat old executable: %w", err)
	}

	// Ensure the new binary has the same permissions
	if err := os.Chmod(newPath, info.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// On Windows, we can't replace a running executable directly.
	// We need to rename it first, then copy the new one.
	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(oldPath, newPath)
	}

	// On Unix systems, we can rename atomically
	return replaceExecutableUnix(oldPath, newPath)
}

// replaceExecutableUnix replaces the executable using atomic rename.
func replaceExecutableUnix(oldPath, newPath string) error {
	// Copy to a temp file in the same directory (for atomic rename)
	dir := filepath.Dir(oldPath)
	tmpFile := filepath.Join(dir, fmt.Sprintf(".stackeye-upgrade-%d", time.Now().UnixNano()))

	// Copy the new binary to the temp location
	if err := copyFile(newPath, tmpFile); err != nil {
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Get original permissions and apply to temp file
	info, err := os.Stat(oldPath)
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to stat original executable: %w", err)
	}
	if err := os.Chmod(tmpFile, info.Mode()); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, oldPath); err != nil {
		os.Remove(tmpFile) // Clean up on failure
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	return nil
}

// replaceExecutableWindows handles Windows-specific executable replacement.
func replaceExecutableWindows(oldPath, newPath string) error {
	// On Windows, rename the old executable to a backup
	backupPath := oldPath + ".old"

	// Remove any existing backup
	_ = os.Remove(backupPath)

	// Rename current to backup
	if err := os.Rename(oldPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup old executable: %w", err)
	}

	// Copy new to old location
	if err := copyFile(newPath, oldPath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, oldPath)
		return fmt.Errorf("failed to copy new executable: %w", err)
	}

	// Remove backup (may fail if still in use, that's OK)
	_ = os.Remove(backupPath)

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// formatSize formats a size in bytes to a human-readable string.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// byteProgressBar provides a simple byte-based progress indicator.
type byteProgressBar struct {
	total     int64
	current   int64
	width     int
	writer    io.Writer
	startTime time.Time
	ticker    *time.Ticker
	done      chan struct{}
	mu        sync.Mutex
}

// newByteProgressBar creates a progress bar for byte-based downloads.
func newByteProgressBar(total int64) *byteProgressBar {
	return &byteProgressBar{
		total:  total,
		width:  30,
		writer: os.Stderr,
		done:   make(chan struct{}),
	}
}

// Start begins rendering the progress bar.
func (b *byteProgressBar) Start() {
	b.mu.Lock()
	b.startTime = time.Now()
	b.ticker = time.NewTicker(100 * time.Millisecond)
	b.mu.Unlock()

	go func() {
		for {
			select {
			case <-b.done:
				return
			case <-b.ticker.C:
				b.render()
			}
		}
	}()
}

// Update sets the current progress.
func (b *byteProgressBar) Update(current int64) {
	b.mu.Lock()
	b.current = current
	b.mu.Unlock()
}

// Finish stops the progress bar and prints a newline.
func (b *byteProgressBar) Finish() {
	b.mu.Lock()
	if b.ticker != nil {
		b.ticker.Stop()
	}
	close(b.done)
	b.mu.Unlock()

	// Final render and newline
	b.render()
	fmt.Fprintln(b.writer)
}

// render draws the progress bar.
func (b *byteProgressBar) render() {
	b.mu.Lock()
	current := b.current
	total := b.total
	writer := b.writer
	width := b.width
	startTime := b.startTime
	b.mu.Unlock()

	if total <= 0 {
		return
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	// Calculate speed
	elapsed := time.Since(startTime)
	var speedStr string
	if elapsed > time.Second && current > 0 {
		bytesPerSec := float64(current) / elapsed.Seconds()
		speedStr = fmt.Sprintf(" %s/s", formatSize(int64(bytesPerSec)))
	}

	fmt.Fprintf(writer, "\r\033[K[%s] %s/%s (%.0f%%)%s",
		bar, formatSize(current), formatSize(total), percent*100, speedStr)
}
