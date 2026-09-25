# PowerShell Installer for Kawaii-Wify on Windows
$ErrorActionPreference = "Stop"

$repo = "obliviousorion/Kawaii-wify"
$asset = "kawaii-wify-windows-amd64.exe"
$downloadUrl = "https://github.com/$repo/releases/latest/download/$asset"

$installDir = "$env:LOCALAPPDATA\kawaii-wify"
$targetExe = "$installDir\kawaii-wify.exe"

Write-Host "Creating installation directory: $installDir"
if (!(Test-Path -Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Write-Host "Downloading $asset from GitHub..."
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $targetExe -UseBasicParsing
} catch {
    Write-Error "Failed to download $asset from $downloadUrl. Ensure a GitHub release is published."
    exit 1
}

Write-Host "Successfully downloaded kawaii-wify.exe"

# Add installDir to User PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to user PATH environment variable..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "PATH updated. You can now run 'kawaii-wify' in any new terminal session."
} else {
    Write-Host "$installDir is already in your PATH."
}

Write-Host "Installation complete! Run 'kawaii-wify --help' to get started."
