$ErrorActionPreference = "Stop"

# Installation directory
$installDir = "$env:USERPROFILE\.nvx\bin"

# Direct download URL (always points to the latest release asset)
$nvxUrl = "https://github.com/walter2310/nvx/releases/latest/download/nvx.exe"
$nvxPath = Join-Path $installDir "nvx.exe"

Write-Host "Downloading nvx from $nvxUrl..."

# Create folder if it doesn't exist
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# Download nvx.exe
Invoke-WebRequest -Uri $nvxUrl -OutFile $nvxPath

# Verify download
if (!(Test-Path $nvxPath)) {
    Write-Error "Failed to download nvx.exe"
    exit 1
}

Write-Host "✅ nvx.exe installed in $installDir"

# Add to PATH if not already there
$userPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to user PATH..."
    setx PATH "$userPath;$installDir" | Out-Null
    Write-Host "Close and reopen your terminal to apply the changes."
} else {
    Write-Host "ℹ$installDir is already in PATH"
}

Write-Host "`n nvx installed successfully. Run 'nvx --help' to get started."
