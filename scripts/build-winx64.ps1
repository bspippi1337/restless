# scripts/build-winx64.ps1
[CmdletBinding()]
param(
  [switch]$SkipTests,
  [switch]$Zip
)

$ErrorActionPreference = "Stop"

function Say($msg) { Write-Host $msg }
function Ok($msg)  { Write-Host "[ OK ] $msg" -ForegroundColor Green }
function Warn($msg){ Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Fail($msg){ Write-Host "[FAIL] $msg" -ForegroundColor Red }

function Have($cmd) {
  return [bool](Get-Command $cmd -ErrorAction SilentlyContinue)
}

function Run-Step {
  param(
    [string]$Name,
    [scriptblock]$Block,
    [switch]$SoftFail
  )
  Say "`n==> $Name"
  try {
    & $Block
    Ok $Name
  } catch {
    if ($SoftFail) {
      Warn "$Name failed (continuing): $($_.Exception.Message)"
    } else {
      throw
    }
  }
}

# Repo root
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $RepoRoot

# Ensure dist folder
$OutDir = Join-Path $RepoRoot "dist\winx64"
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

Run-Step "Environment" {
  $arch = $env:PROCESSOR_ARCHITECTURE
  Say "RepoRoot: $RepoRoot"
  Say "Arch: $arch"
  if ($arch -notin @("AMD64","ARM64")) { Warn "Unknown arch: $arch (expected AMD64)" }
}

# --- Go build ---
if (-not (Have "go")) {
  Fail "Go not found. If you use Chocolatey:  choco install -y golang"
  exit 1
}

Run-Step "Go: go version" { go version | Out-Host }

Run-Step "Go: go mod tidy (soft)" { go mod tidy } -SoftFail

if (-not $SkipTests) {
  Run-Step "Go: tests (soft)" {
    $env:CGO_ENABLED="0"
    go test ./... -count=1 -tags "netgo osusergo"
  } -SoftFail
} else {
  Warn "Skipping tests: -SkipTests"
}

Run-Step "Go: build restless.exe" {
  $env:CGO_ENABLED="0"
  $exe = Join-Path $OutDir "restless.exe"
  go build -trimpath -tags "netgo osusergo" -ldflags "-s -w" -o $exe .\cmd\restless
  if (-not (Test-Path $exe)) { throw "Build did not produce $exe" }
}

# --- C core build (optional) ---
$CoreDir = Join-Path $RepoRoot "corec"
if (Test-Path $CoreDir) {
  Run-Step "C core: build restless-core.exe (no vendored deps)" {
    $srcGlob = Join-Path $CoreDir "src\*.c"
    $incDir  = Join-Path $CoreDir "include"
    $outExe  = Join-Path $OutDir "restless-core.exe"

    # Prefer MSVC if available; otherwise MinGW gcc if available.
    if (Have "cl") {
      # cl.exe must be on PATH (Developer PowerShell / VsDevCmd).
      & cl /nologo /O2 /W3 /std:c11 /I"$incDir" $srcGlob /Fe:"$outExe" | Out-Host
    }
    elseif (Have "gcc") {
      & gcc -O2 -Wall -Wextra -std=c11 -I"$incDir" $srcGlob -o "$outExe" | Out-Host
    }
    else {
      throw "No C compiler found. Install Visual Studio Build Tools (cl) or MinGW (gcc)."
    }

    if (-not (Test-Path $outExe)) { throw "C build did not produce $outExe" }
  }
} else {
  Warn "corec/ not found; skipping C core build."
}

# --- Optional release zip ---
if ($Zip) {
  Run-Step "Zip: dist/winx64 -> dist/restless-winx64.zip" {
    $zipPath = Join-Path $RepoRoot "dist\restless-winx64.zip"
    if (Test-Path $zipPath) { Remove-Item -Force $zipPath }
    Compress-Archive -Path (Join-Path $OutDir "*") -DestinationPath $zipPath -Force
    Say "ZIP: $zipPath"
  }
}

Say "`n✅ Done."
Say "Artifacts:"
Say "  - $OutDir\restless.exe"
if (Test-Path (Join-Path $OutDir "restless-core.exe")) {
  Say "  - $OutDir\restless-core.exe"
}
