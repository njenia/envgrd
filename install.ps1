# PowerShell installation script for envgrd
# Usage: Invoke-WebRequest -UseBasicParsing https://raw.githubusercontent.com/njenia/envgrd/main/install.ps1 | Invoke-Expression

$ErrorActionPreference = "Continue"  # Changed to Continue so we can see errors

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
        Write-Output ""
        Write-Output "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        return
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
            Write-Output "Downloading from: $DOWNLOAD_URL"
            Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $ZIP_PATH -UseBasicParsing
            Write-Output "Download complete: $ZIP_PATH"
        } catch {
            Write-ColorOutput Red "Error: Failed to download ${BINARY_NAME}"
            Write-Output "URL: $DOWNLOAD_URL"
            Write-Output "Error: $_"
            Write-Output ""
            Write-Output "Press any key to exit..."
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
            return
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
            Write-Output "Extracted to: $EXTRACT_DIR"
            Write-Output "Contents:"
            Get-ChildItem -Path $EXTRACT_DIR -Recurse | Select-Object FullName | ForEach-Object { Write-Output "  $($_.FullName)" }
            Write-Output ""
            Write-Output "Press any key to exit..."
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
            return
        }
        
        Write-Output "Found binary at: $($BINARY_PATH.FullName)"
        
        # Create install directory
        Write-Output "Install directory: $INSTALL_DIR"
        if (-not (Test-Path $INSTALL_DIR)) {
            Write-Output "Creating install directory..."
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
            if (-not (Test-Path $INSTALL_DIR)) {
                Write-ColorOutput Red "Error: Failed to create install directory: $INSTALL_DIR"
                Write-Output ""
                Write-Output "Press any key to exit..."
                $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
                return
            }
        }
        
        # Install
        Write-Output "Installing to $INSTALL_DIR..."
        $INSTALL_PATH = Join-Path $INSTALL_DIR "${BINARY_NAME}.exe"
        Write-Output "Copying from: $($BINARY_PATH.FullName)"
        Write-Output "Copying to: $INSTALL_PATH"
        try {
            Copy-Item -Path $BINARY_PATH.FullName -Destination $INSTALL_PATH -Force
            Write-Output "Copy successful"
        } catch {
            Write-ColorOutput Red "Error: Failed to copy binary"
            Write-Output "Error: $_"
            Write-Output ""
            Write-Output "Press any key to exit..."
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
            return
        }
        
        # Verify file exists
        if (-not (Test-Path $INSTALL_PATH)) {
            Write-ColorOutput Red "Error: Binary was not copied successfully"
            Write-Output "Expected at: $INSTALL_PATH"
            Write-Output ""
            Write-Output "Press any key to exit..."
            $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
            return
        }
        Write-Output "Binary verified at: $INSTALL_PATH"
        
        # Add to PATH if not already there
        Write-Output "Checking PATH..."
        $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($CurrentPath -notlike "*$INSTALL_DIR*") {
            Write-Output "Adding $INSTALL_DIR to PATH..."
            $NewPath = $CurrentPath + ";$INSTALL_DIR"
            try {
                [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
                $env:Path += ";$INSTALL_DIR"
                Write-Output "PATH updated successfully"
            } catch {
                Write-ColorOutput Yellow "Warning: Failed to update PATH automatically"
                Write-Output "Error: $_"
                Write-Output "Please add $INSTALL_DIR to your PATH manually"
            }
        } else {
            Write-Output "PATH already contains $INSTALL_DIR"
        }
        
        # Verify installation
        Write-Output ""
        Write-Output "Verifying installation..."
        Write-Output "Install path: $INSTALL_PATH"
        Write-Output "File exists: $(Test-Path $INSTALL_PATH)"
        
        if (Test-Path $INSTALL_PATH) {
            $FileInfo = Get-Item $INSTALL_PATH
            Write-Output "File size: $($FileInfo.Length) bytes"
            Write-Output "File date: $($FileInfo.LastWriteTime)"
        }
        
        Start-Sleep -Seconds 1  # Give PATH a moment to update
        
        if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
            try {
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
            } catch {
                Write-ColorOutput Green "✓ Successfully installed ${BINARY_NAME}"
                Write-Output "Location: $INSTALL_PATH"
                Write-ColorOutput Yellow "Note: You may need to restart your terminal for PATH changes to take effect"
            }
        } else {
            Write-ColorOutput Green "✓ Successfully installed ${BINARY_NAME}"
            Write-Output "Location: $INSTALL_PATH"
            Write-ColorOutput Yellow "Note: You may need to restart your terminal for PATH changes to take effect"
            Write-Output ""
            Write-Output "To use it now, run:"
            Write-Output "  & `"$INSTALL_PATH`" scan"
        }
        
        Write-Output ""
        Write-Output "Installation complete!"
    } finally {
        # Cleanup
        Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Run installation
try {
    Install-Envgrd
} catch {
    Write-ColorOutput Red "Fatal error during installation:"
    Write-Output $_
    Write-Output $_.ScriptStackTrace
}

Write-Output ""
Write-Output "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

