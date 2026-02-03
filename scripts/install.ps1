# install.ps1 - Install StackEye CLI on Windows
#
# This script downloads and installs the StackEye CLI binary for Windows.
# It auto-detects architecture, verifies checksums, and installs the binary.
# Downloads from CloudFlare CDN (releases.stackeye.io) with GitHub fallback.
#
# Usage:
#   iwr -useb https://releases.stackeye.io/install.ps1 | iex
#   $env:STACKEYE_VERSION = "v1.0.0"; iwr -useb https://releases.stackeye.io/install.ps1 | iex
#   $env:STACKEYE_INSTALL_DIR = "C:\custom\path"; iwr -useb https://releases.stackeye.io/install.ps1 | iex
#
# Environment variables:
#   STACKEYE_VERSION      - Install specific version (default: latest)
#   STACKEYE_INSTALL_DIR  - Installation directory (default: $env:LOCALAPPDATA\StackEye\bin)
#   STACKEYE_NO_VERIFY    - Skip checksum verification (not recommended)
#
# Supported platforms:
#   - Windows 10/11 (x64)
#   - Windows Server 2019+ (x64)
#
# Requires PowerShell 5.1 or later.

#Requires -Version 5.1

$ErrorActionPreference = "Stop"

# Configuration
$GithubRepo = "StackEye-IO/stackeye-cli"
$BinaryName = "stackeye"

# Distribution URLs (CDN primary, GitHub fallback)
$CdnBaseUrl = "https://releases.stackeye.io/cli"
$GithubBaseUrl = "https://github.com/$GithubRepo/releases/download"

# --- Output helpers ---

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Blue -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
}

# --- Platform detection ---

function Get-Architecture {
    # Check .NET RuntimeInformation first (PowerShell 6+)
    try {
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
        switch ($arch) {
            "X64"  { return "amd64" }
            "Arm64" { return "arm64" }
            default {
                Write-Err "Unsupported architecture: $arch"
                Write-Err "StackEye CLI supports x64 (amd64) on Windows."
                exit 1
            }
        }
    } catch {
        # Fallback for PowerShell 5.1
    }

    # Fallback: environment variable
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Err "Unsupported architecture: $arch"
            Write-Err "StackEye CLI supports x64 (amd64) on Windows."
            exit 1
        }
    }
}

function Test-WindowsVersion {
    $os = [System.Environment]::OSVersion
    if ($os.Platform -ne "Win32NT") {
        Write-Err "This installer is for Windows only."
        Write-Err "For Linux/macOS, use: curl -fsSL https://releases.stackeye.io/install.sh | bash"
        exit 1
    }

    # Windows 10 = 10.0.10240+, Server 2019 = 10.0.17763+
    if ($os.Version.Major -lt 10) {
        Write-Warn "Windows version $($os.Version) detected. This installer targets Windows 10+ / Server 2019+."
    }
}

# --- Version resolution ---

function Get-LatestVersion {
    $url = "https://api.github.com/repos/$GithubRepo/releases/latest"

    try {
        # Use TLS 1.2 (required by GitHub API)
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $response = Invoke-RestMethod -Uri $url -ErrorAction Stop
        $tag = $response.tag_name
        if ([string]::IsNullOrEmpty($tag)) {
            throw "Empty tag_name in API response"
        }
        return $tag
    } catch {
        Write-Err "Failed to determine latest version from GitHub."
        Write-Err "  Error: $($_.Exception.Message)"
        Write-Err "Please check your internet connection or set `$env:STACKEYE_VERSION."
        exit 1
    }
}

# --- Download helpers ---

# Suppress progress bar for faster downloads in PowerShell 5.1
$ProgressPreference = 'SilentlyContinue'

function Invoke-DownloadSilent {
    param(
        [string]$Url,
        [string]$OutFile
    )

    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing -ErrorAction Stop
        return $true
    } catch {
        return $false
    }
}

function Invoke-DownloadWithFallback {
    param(
        [string]$CdnUrl,
        [string]$GithubUrl,
        [string]$OutFile
    )

    Write-Info "Downloading from CDN..."
    if (Invoke-DownloadSilent -Url $CdnUrl -OutFile $OutFile) {
        return
    }

    Write-Warn "CDN download failed, trying GitHub..."
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Write-Info "Downloading: $GithubUrl"
        Invoke-WebRequest -Uri $GithubUrl -OutFile $OutFile -UseBasicParsing -ErrorAction Stop
    } catch {
        Write-Err "Failed to download from both CDN and GitHub."
        Write-Err "  CDN URL: $CdnUrl"
        Write-Err "  GitHub URL: $GithubUrl"
        Write-Err "  Error: $($_.Exception.Message)"
        exit 1
    }
}

# --- Checksum verification ---

function Test-Checksum {
    param(
        [string]$FilePath,
        [string]$ExpectedHash
    )

    $actual = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
    $expected = $ExpectedHash.ToLower()

    if ($actual -ne $expected) {
        Write-Err "Checksum verification failed!"
        Write-Err "  Expected: $expected"
        Write-Err "  Actual:   $actual"
        return $false
    }

    Write-Info "Checksum verified successfully"
    return $true
}

function Get-ExpectedChecksum {
    param(
        [string]$ChecksumsFile,
        [string]$ArchiveName
    )

    $lines = Get-Content -Path $ChecksumsFile
    $escapedName = [regex]::Escape($ArchiveName)
    foreach ($line in $lines) {
        if ($line -match "^\s*(\S+)\s+.*${escapedName}\s*$") {
            return $Matches[1].Trim()
        }
    }
    return $null
}

# --- Installation ---

function Get-InstallDir {
    # Check for override
    if (-not [string]::IsNullOrEmpty($env:STACKEYE_INSTALL_DIR)) {
        return $env:STACKEYE_INSTALL_DIR
    }

    # Default: user-level directory (no admin required)
    return Join-Path $env:LOCALAPPDATA "StackEye\bin"
}

function Test-InPath {
    param([string]$Dir)

    $pathDirs = $env:PATH -split ';'
    foreach ($d in $pathDirs) {
        if ($d.TrimEnd('\') -ieq $Dir.TrimEnd('\')) {
            return $true
        }
    }
    return $false
}

function Add-ToUserPath {
    param([string]$Dir)

    # Get the current User PATH from the registry (persistent)
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ([string]::IsNullOrEmpty($currentPath)) {
        $currentPath = ""
    }

    # Check if already in user PATH
    $pathDirs = $currentPath -split ';' | Where-Object { $_ -ne '' }
    foreach ($d in $pathDirs) {
        if ($d.TrimEnd('\') -ieq $Dir.TrimEnd('\')) {
            return  # Already in PATH
        }
    }

    # Append to user PATH
    $newPath = if ([string]::IsNullOrEmpty($currentPath)) {
        $Dir
    } else {
        "$currentPath;$Dir"
    }
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

    # Also update current session PATH
    if (-not (Test-InPath -Dir $Dir)) {
        $env:PATH = "$env:PATH;$Dir"
    }

    Write-Info "Added $Dir to your User PATH (persistent across sessions)"
}

function Install-Binary {
    param(
        [string]$SourcePath,
        [string]$InstallDir
    )

    # Create directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $destPath = Join-Path $InstallDir "$BinaryName.exe"

    # Copy binary
    Write-Info "Installing to $destPath"
    Copy-Item -Path $SourcePath -Destination $destPath -Force

    Write-Success "StackEye CLI installed to $destPath"

    # Add to PATH if not present
    if (-not (Test-InPath -Dir $InstallDir)) {
        Add-ToUserPath -Dir $InstallDir
    }
}

# --- Main ---

function Main {
    Write-Host ""
    Write-Host "  StackEye CLI Installer (Windows)"
    Write-Host "  ================================="
    Write-Host ""

    # Validate platform
    Test-WindowsVersion

    # Detect architecture
    $arch = Get-Architecture
    Write-Info "Detected architecture: $arch"

    # GoReleaser excludes windows/arm64
    if ($arch -eq "arm64") {
        Write-Err "StackEye CLI does not currently provide Windows ARM64 builds."
        Write-Err "Please use the x64 build under Windows ARM64 emulation, or build from source."
        exit 1
    }

    # Determine version
    $version = $env:STACKEYE_VERSION
    if ([string]::IsNullOrEmpty($version)) {
        Write-Info "Determining latest version..."
        $version = Get-LatestVersion
    }
    Write-Info "Version: $version"

    # Strip leading 'v' for archive name
    $versionNoV = $version.TrimStart('v')

    # Construct archive name (matches GoReleaser output)
    $archiveName = "${BinaryName}_${versionNoV}_windows_${arch}.zip"
    $checksumsName = "checksums.txt"

    # Construct download URLs
    $cdnUrl = "$CdnBaseUrl/$version"
    $githubUrl = "$GithubBaseUrl/$version"

    # Create temporary directory
    $tmpDir = Join-Path $env:TEMP "stackeye-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        $archivePath = Join-Path $tmpDir $archiveName
        $checksumsPath = Join-Path $tmpDir $checksumsName

        # Download checksums
        Invoke-DownloadWithFallback `
            -CdnUrl "$cdnUrl/$checksumsName" `
            -GithubUrl "$githubUrl/$checksumsName" `
            -OutFile $checksumsPath

        # Download archive
        Invoke-DownloadWithFallback `
            -CdnUrl "$cdnUrl/$archiveName" `
            -GithubUrl "$githubUrl/$archiveName" `
            -OutFile $archivePath

        # Verify checksum
        if ([string]::IsNullOrEmpty($env:STACKEYE_NO_VERIFY)) {
            $expectedHash = Get-ExpectedChecksum -ChecksumsFile $checksumsPath -ArchiveName $archiveName
            if ([string]::IsNullOrEmpty($expectedHash)) {
                Write-Err "Could not find checksum for $archiveName"
                exit 1
            }

            if (-not (Test-Checksum -FilePath $archivePath -ExpectedHash $expectedHash)) {
                exit 1
            }
        } else {
            Write-Warn "Checksum verification skipped (STACKEYE_NO_VERIFY is set)"
        }

        # Extract archive
        Write-Info "Extracting archive..."
        $extractDir = Join-Path $tmpDir "extracted"
        Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

        # Find the binary
        $binaryPath = Join-Path $extractDir "$BinaryName.exe"
        if (-not (Test-Path $binaryPath)) {
            Write-Err "Binary not found in archive at expected path: $binaryPath"
            exit 1
        }

        # Determine install directory
        $installDir = Get-InstallDir

        # Install
        Install-Binary -SourcePath $binaryPath -InstallDir $installDir

        # Verify installation
        Write-Host ""
        $exePath = Join-Path $installDir "$BinaryName.exe"
        if (Test-Path $exePath) {
            Write-Info "Verifying installation..."
            try {
                & $exePath version
            } catch {
                Write-Warn "Could not run version check: $($_.Exception.Message)"
            }
        }

        Write-Host ""
        Write-Success "StackEye CLI has been installed successfully!"
        Write-Host ""
        Write-Host "  Get started:"
        Write-Host "    stackeye login    # Authenticate with StackEye"
        Write-Host "    stackeye --help   # See all available commands"
        Write-Host ""

        # Remind about new shell if PATH was updated
        if (-not (Test-InPath -Dir $installDir)) {
            Write-Warn "You may need to restart your terminal for PATH changes to take effect."
        }
    } finally {
        # Cleanup
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run
Main
