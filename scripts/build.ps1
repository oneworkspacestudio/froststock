param(
 [string]$Go='go',
 [string]$Python='python',
 [Parameter(Mandatory)][string]$Output
)
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
$safe="safe.directory=$($root.Replace('\','/'))"
$outputPath=$ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Output)
function Check($message){if($LASTEXITCODE){throw $message}}
Push-Location $root
try {
 $status=git -c $safe status --porcelain
 Check 'Git status failed'
 if($status){throw 'Commit the intended source before building'}
 $commit=(git -c $safe rev-parse HEAD).Trim()
 Check 'Git source commit unavailable'
 if(Test-Path -LiteralPath $outputPath){throw 'Use a fresh output directory'}
 $sdk=& $Go list -m -json github.com/oneworkspacestudio/one-workspace | ConvertFrom-Json
 Check 'Workspace AppHost module unavailable'
 if(!$sdk.Version -or $sdk.Replace){throw 'Use a published Workspace AppHost source pin'}
 & $Go test -buildvcs=false ./...
 Check 'Froststock tests failed'
 & $Go vet ./...
 Check 'Froststock vet failed'
 node scripts/froststock-recognition-test.mjs
 Check 'Froststock recognition tests failed'
 New-Item -ItemType Directory -Path $outputPath | Out-Null
 $platform=@{GOOS=$env:GOOS; GOARCH=$env:GOARCH; CGO_ENABLED=$env:CGO_ENABLED}
 try {
  $env:GOOS='windows'
  $env:GOARCH='amd64'
  $env:CGO_ENABLED='0'
  & $Go build -buildvcs=false -trimpath -ldflags '-s -w' -o "$outputPath/Froststock.exe" ./cmd/froststock
  Check 'Froststock Windows build failed'
 } finally {
  foreach($entry in $platform.GetEnumerator()){[Environment]::SetEnvironmentVariable($entry.Key,$entry.Value,'Process')}
 }
 Copy-Item -LiteralPath LICENSE,README.md -Destination $outputPath
 Copy-Item -LiteralPath internal/froststock/NOTICE -Destination "$outputPath/NOTICE"
 $contents=[Text.StringBuilder]::new()
 $modules=& $Go list -buildvcs=false -deps -f '{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}' ./cmd/froststock | Where-Object {$_} | Sort-Object -Unique
 Check 'Go dependency inventory failed'
 foreach($row in $modules){
  $parts=$row.Split('|',3)
  if($parts[0] -eq 'github.com/oneworkspacestudio/froststock'){continue}
  $notices=@(Get-ChildItem -LiteralPath $parts[2] -File | Where-Object Name -Match '^(LICENSE|LICENCE|COPYING|NOTICE)(\..*)?$')
  if(!$notices.Count){throw ('Missing module licence for '+$parts[0])}
  [void]$contents.AppendLine("===== $($parts[0]) $($parts[1]) =====")
  foreach($notice in $notices){[void]$contents.AppendLine([IO.File]::ReadAllText($notice.FullName))}
 }
 $goRoot=(& $Go env GOROOT).Trim()
 Check 'Go runtime path unavailable'
 [void]$contents.AppendLine('===== Go runtime =====').AppendLine([IO.File]::ReadAllText((Join-Path $goRoot 'LICENSE')))
 foreach($notice in Get-ChildItem -LiteralPath internal/froststock/web/assets -Recurse -File | Where-Object Name -Match 'LICEN[CS]E|NOTICE|^SOURCES\.md$'){
  [void]$contents.AppendLine("===== $([IO.Path]::GetRelativePath($root,$notice.FullName)) =====").AppendLine([IO.File]::ReadAllText($notice.FullName))
 }
 [IO.File]::WriteAllText("$outputPath/THIRD-PARTY-LICENSES.txt",$contents.ToString())
 git -c $safe archive --format=zip --prefix=froststock/ "--output=$outputPath/froststock-source.zip" $commit
 Check 'Source archive failed'
 & $Python scripts/prepare-release.py $outputPath $commit $sdk.Version
 Check 'Froststock manifest or checksums failed'
} finally { Pop-Location }
