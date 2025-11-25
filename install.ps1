# PowerShell installation script for envgrd
# Usage: Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/njenia/envgrd/main/install.ps1 | Invoke-Expression

$ErrorActionPreference = "Stop"

# GitHub repository
$REPO = "njenia/envgrd"
$BINARY_NAME = "envgrd"
$INSTALL_DIR = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\envgrd" }

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

# Detect platform
function Detect-Platform {
    $OS = "windows"
    $ARCH = $env:PROCESSOR_ARCHITECTURE
    
    # Map architecture
    switch ($ARCH) {
        "AMD64" { $ARCH = "amd64" }
        "ARM64" { $ARCH = "arm64" }
        default {
            Write-ColorOutput Red "Error: Unsupported architecture: $ARCH"
            exit 1
        }
    }
    
    # Windows only supports amd64 in releases
    if ($ARCH -ne "amd64") {
        Write-ColorOutput Yellow "Note: ${ARCH} Windows binaries are not available in releases."
        Write-Output "Please build from source:"
        Write-Output "  git clone https://github.com/$REPO.git"
        Write-Output "  cd envgrd && make build"
        exit 0
    }
    
    $PLATFORM = "windows"
    $EXT = "zip"
    
    return @{
        Platform = $PLATFORM
        Arch = $ARCH
        Ext = $EXT
    }
}

# Download and install
function Install-Envgrd {
    $platformInfo = Detect-Platform
    $PLATFORM = $platformInfo.Platform
    $ARCH = $platformInfo.Arch
    $EXT = $platformInfo.Ext
    
    $VERSION = if ($env:VERSION) { $env:VERSION } else { "latest" }
    
    if ($VERSION -eq "latest") {
        $DOWNLOAD_URL = "https://github.com/$REPO/releases/latest/download/${BINARY_NAME}-${PLATFORM}-${ARCH}.${EXT}"
    } else {
        $DOWNLOAD_URL = "https://github.com/$REPO/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}-${ARCH}.${EXT}"
    }
    
    Write-ColorOutput Green "Installing ${BINARY_NAME}..."
    Write-Output "Platform: ${PLATFORM}-${ARCH}"
    Write-Output "Download URL: $DOWNLOAD_URL"
    
    # Create temporary directory
    $TMP_DIR = Join-Path $env:TEMP "envgrd-install-$(New-Guid)"
    New-Item -ItemType Directory -Path $TMP_DIR -Force | Out-Null
    
    try {
        # Download
        Write-Output "Downloading..."
        $ZIP_PATH = Join-Path $TMP_DIR "${BINARY_NAME}.${EXT}"
        
        try {
            Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $ZIP_PATH -UseBasicParsing
        } catch {
            Write-ColorOutput Red "Error: Failed to download ${BINARY_NAME}"
            Write-Output "URL: $DOWNLOAD_URL"
            Write-Output "Error: $_"
            exit 1
        }
        
        # Extract
        Write-Output "Extracting..."
        $EXTRACT_DIR = Join-Path $TMP_DIR "extract"
        New-Item -ItemType Directory -Path $EXTRACT_DIR -Force | Out-Null
        Expand-Archive -Path $ZIP_PATH -DestinationPath $EXTRACT_DIR -Force
        
        # Find the binary (could be envgrd.exe or in a subdirectory)
        $BINARY_PATH = Get-ChildItem -Path $EXTRACT_DIR -Filter "${BINARY_NAME}.exe" -Recurse | Select-Object -First 1
        
        if (-not $BINARY_PATH) {
            Write-ColorOutput Red "Error: Could not find ${BINARY_NAME}.exe in downloaded archive"
            exit 1
        }
        
        # Create install directory
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }
        
        # Install
        Write-Output "Installing to $INSTALL_DIR..."
        $INSTALL_PATH = Join-Path $INSTALL_DIR "${BINARY_NAME}.exe"
        Copy-Item -Path $BINARY_PATH.FullName -Destination $INSTALL_PATH -Force
        
        # Add to PATH if not already there
        $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($CurrentPath -notlike "*$INSTALL_DIR*") {
            Write-Output "Adding $INSTALL_DIR to PATH..."
            $NewPath = $CurrentPath + ";$INSTALL_DIR"
            [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
            $env:Path += ";$INSTALL_DIR"
        }
        
        # Verify installation
        Write-Output "Verifying installation..."
        Start-Sleep -Seconds 1  # Give PATH a moment to update
        
        if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
            $INSTALLED_VERSION = & $BINARY_NAME version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-ColorOutput Green "✓ Successfully installed ${BINARY_NAME}"
                Write-Output "Version: $INSTALLED_VERSION"
                Write-Output "Run '${BINARY_NAME} scan' to get started!"
            } else {
                Write-ColorOutput Green "✓ Successfully installed ${BINARY_NAME}"
                Write-Output "Location: $INSTALL_PATH"
                Write-ColorOutput Yellow "Note: You may need to restart your terminal for PATH changes to take effect"
            }
        } else {
            Write-ColorOutput Green "✓ Successfully installed ${BINARY_NAME}"
            Write-Output "Location: $INSTALL_PATH"
            Write-ColorOutput Yellow "Note: You may need to restart your terminal for PATH changes to take effect"
            Write-Output "Or run: ${INSTALL_PATH} scan"
        }
    } finally {
        # Cleanup
        Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Run installation
Install-Envgrd

