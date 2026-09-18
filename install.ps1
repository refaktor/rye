<#
.SYNOPSIS
  Install Rye on Windows (PowerShell).

.DESCRIPTION
  Downloads the latest (or specified) Rye release for Windows from GitHub
  and installs rye.exe into a chosen directory (default: Program Files\Rye if admin,
  otherwise %LOCALAPPDATA%\Programs\Rye). Optionally adds that directory to the
  current user's PATH.

.PARAMETER Version
  Optional git tag to install, e.g. v0.10.3. If omitted, installs latest release.

.PARAMETER InstallDir
  Optional install directory. If omitted, uses Program Files if elevated, otherwise
  %LOCALAPPDATA%\Programs\Rye.

.PARAMETER AddToPath
  If set, the install directory is added to the current user's PATH (persistent) if
  not already present.

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File .\install.ps1 -AddToPath

.EXAMPLE
  iwr https://raw.githubusercontent.com/refaktor/rye/main/install.ps1 -UseBasicParsing | iex
#>

[CmdletBinding()]
param(
  [string]$Version,
  [string]$InstallDir,
  [switch]$AddToPath
)

$ErrorActionPreference = 'Stop'

function Test-Admin {
  $wid = [System.Security.Principal.WindowsIdentity]::GetCurrent()
  $prp = New-Object System.Security.Principal.WindowsPrincipal($wid)
  return $prp.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Get-ArchString {
  # Map to goreleaser archive name format (x86_64 | arm64)
  $arch = $env:PROCESSOR_ARCHITECTURE
  switch -Regex ($arch) {
    '^(AMD64|X86_64)$' { return 'x86_64' }
    '^(ARM64)$'        { return 'arm64' }
    default { throw "Unsupported architecture: $arch" }
  }
}

function Get-LatestTag {
  Write-Verbose 'Fetching latest release tag from GitHub...'
  $api = 'https://api.github.com/repos/refaktor/rye/releases/latest'
  $tag = try {
    (Invoke-RestMethod -UseBasicParsing -Uri $api -Headers @{ 'User-Agent' = 'rye-installer' }).tag_name
  } catch {
    throw "Failed to query latest release: $($_.Exception.Message)"
  }
  if ([string]::IsNullOrWhiteSpace($tag)) { throw 'Could not determine latest release version.' }
  return $tag
}

function Ensure-Dir([string]$dir) {
  if (-not (Test-Path -LiteralPath $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }
}

function Add-ToUserPath([string]$dir) {
  $current = [Environment]::GetEnvironmentVariable('PATH', 'User')
  $paths = ($current -split ';') | Where-Object { $_ -and $_.Trim() }
  if ($paths -notcontains $dir) {
    $newPath = if ($current) { $current.TrimEnd(';') + ";$dir" } else { $dir }
    [Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')
    Write-Host "Added to user PATH: $dir"
    Write-Host 'You may need to open a new terminal for PATH changes to take effect.'
  } else {
    Write-Verbose 'Install directory already on PATH.'
  }
}

# 1) Detect arch and OS
if ($IsWindows -ne $true) { throw 'This installer supports Windows only.' }
$archStr = Get-ArchString

# 2) Resolve version and asset name
$tag = if ($Version) { $Version } else { Get-LatestTag }
$asset = "rye_Windows_${archStr}.zip"
$downloadUrl = "https://github.com/refaktor/rye/releases/download/$tag/$asset"

# 3) Pick install directory
$defaultDir = if (Test-Admin) { Join-Path $env:ProgramFiles 'Rye' } else { Join-Path $env:LOCALAPPDATA 'Programs\Rye' }
$targetDir = if ($InstallDir) { $InstallDir } else { $defaultDir }
Ensure-Dir $targetDir

# 4) Download and extract to temp
$work = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "rye-install-" + [System.Guid]::NewGuid())) -Force
try {
  $zipPath = Join-Path $work.FullName $asset
  Write-Host "Downloading $tag ($asset) ..."
  try {
    Invoke-WebRequest -UseBasicParsing -Uri $downloadUrl -OutFile $zipPath
  } catch {
    throw "Download failed from $downloadUrl: $($_.Exception.Message)"
  }

  Write-Host 'Extracting...'
  Expand-Archive -Path $zipPath -DestinationPath $work.FullName -Force

  # Find rye.exe in extracted content
  $exe = Get-ChildItem -Path $work.FullName -Filter 'rye.exe' -Recurse | Select-Object -First 1
  if (-not $exe) { throw 'rye.exe not found in archive.' }

  $destExe = Join-Path $targetDir 'rye.exe'
  Write-Host "Installing to: $destExe"
  if (Test-Path -LiteralPath $destExe) { Remove-Item -LiteralPath $destExe -Force }
  Move-Item -LiteralPath $exe.FullName -Destination $destExe -Force

  Write-Host ''
  Write-Host "Successfully installed Rye ($tag) to $destExe" -ForegroundColor Green
  if ($AddToPath) { Add-ToUserPath $targetDir }
  Write-Host "Run 'rye' from a new terminal to start the REPL."
}
finally {
  try { Remove-Item -Recurse -Force -LiteralPath $work.FullName | Out-Null } catch { }
}
