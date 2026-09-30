# PowerShell Installer & Updater for Kawaii-Wify on Windows
param (
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$repo = "obliviousorion/Kawaii-wify"
$asset = "kawaii-wify-windows-amd64.exe"
$downloadUrl = "https://github.com/$repo/releases/latest/download/$asset"

$installDir = "$env:LOCALAPPDATA\kawaii-wify"
$targetExe = "$installDir\kawaii-wify.exe"

# 1. Check if already installed and check for updates
$isInstalled = Test-Path -Path $targetExe
$latestTag = $null

if ($isInstalled -and -not $Force) {
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        $apiUrl = "https://api.github.com/repos/$repo/releases/latest"
        $release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing -Headers @{ "User-Agent" = "kawaii-wify-installer" }
        $latestTag = $release.tag_name
    } catch {
        # If GitHub API is rate-limited or unreachable, proceed with download
        $latestTag = $null
    }

    if ($latestTag) {
        try {
            $currentVerOutput = (& $targetExe --version 2>&1) | Out-String
            if ($currentVerOutput -match [regex]::Escape($latestTag)) {
                Write-Host "Kawaii-Wify is already up to date ($latestTag). Nothing to do!" -ForegroundColor Green
                Write-Host "To force reinstall, re-run with -Force flag: irm ... | iex -ArgumentList '-Force'"
                exit 0
            } else {
                Write-Host "Update available! Installing latest release ($latestTag)..." -ForegroundColor Cyan
            }
        } catch {
            Write-Host "Upgrading existing installation..." -ForegroundColor Cyan
        }
    }
}

# 2. Stop running daemon if active to prevent Windows file locking
$wasRunning = $false
$runningProcesses = Get-Process -Name "kawaii-wify" -ErrorAction SilentlyContinue
if ($runningProcesses) {
    Write-Host "Detected active kawaii-wify daemon. Stopping daemon to update binary..." -ForegroundColor Yellow
    $wasRunning = $true
    try {
        if ($isInstalled) {
            & $targetExe stop | Out-Null
        }
    } catch {}
    Start-Sleep -Milliseconds 750
    $remaining = Get-Process -Name "kawaii-wify" -ErrorAction SilentlyContinue
    if ($remaining) {
        $remaining | Stop-Process -Force -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 500
    }
}

# 3. Ensure installation directory exists
if (!(Test-Path -Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# 4. Download latest binary to temporary file first
$tempExe = "$targetExe.new"
Write-Host "Downloading $asset from GitHub..."
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempExe -UseBasicParsing
} catch {
    Write-Error "Failed to download $asset from $downloadUrl. Ensure a GitHub release is published."
    if (Test-Path -Path $tempExe) { Remove-Item -Path $tempExe -Force }
    exit 1
}

# 5. Atomically replace target executable
Move-Item -Path $tempExe -Destination $targetExe -Force
Write-Host "Successfully installed kawaii-wify.exe to $targetExe" -ForegroundColor Green

# 6. Add installDir to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to user PATH environment variable..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "PATH updated. You can now run 'kawaii-wify' in any new terminal session."
}

# 7. Restart daemon if it was previously running
if ($wasRunning) {
    Write-Host "Restarting kawaii-wify daemon..." -ForegroundColor Cyan
    try {
        & $targetExe start | Out-Null
        Write-Host "Daemon successfully restarted!" -ForegroundColor Green
    } catch {
        Write-Warning "Could not automatically restart daemon. Run 'kawaii-wify start' manually."
    }
}

if ($isInstalled) {
    Write-Host "Upgrade complete! Run 'kawaii-wify status' to check connection." -ForegroundColor Green
} else {
    Write-Host "Installation complete! Run 'kawaii-wify --help' to get started." -ForegroundColor Green
}

