param([Parameter(Mandatory)][string]$Output)
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
$safe="safe.directory=$($root.Replace('\','/'))"
$repo='oneworkspacestudio/froststock'
$directory=(Resolve-Path -LiteralPath $Output).Path
$info=Get-Content -LiteralPath "$directory/build-info.json" -Raw | ConvertFrom-Json
$tag=[string]$info.tag
if($info.repository -cne $repo -or $tag -notmatch '^v\d+\.\d+\.\d+(?:-[A-Za-z0-9.-]+)?$'){throw 'Invalid release metadata'}
$dirty=git -c $safe -C $root status --porcelain
if($LASTEXITCODE -or $dirty){throw 'Publish from a clean committed checkout'}
$source=(git -c $safe -C $root rev-parse HEAD).Trim()
if($LASTEXITCODE -or $source -cne $info.sourceCommit){throw 'Release source differs from the checkout'}
$remote=git -c $safe -C $root ls-remote origin refs/heads/main
if($LASTEXITCODE -or $remote -notmatch "^$source\s"){throw 'Push the release source to Froststock main'}
$checksums=@{}
foreach($line in Get-Content -LiteralPath "$directory/SHA256SUMS.txt"){
 if($line -notmatch '^([a-f0-9]{64})  ([A-Za-z0-9_.-]+)$' -or $checksums.ContainsKey($Matches[2])){throw 'Invalid checksum list'}
 $checksums[$Matches[2]]=$Matches[1]
}
$local=@{}
foreach($name in (@($checksums.Keys) + 'SHA256SUMS.txt')){
 $file=Get-Item -LiteralPath (Join-Path $directory $name)
 $digest=(Get-FileHash -LiteralPath $file.FullName).Hash.ToLowerInvariant()
 if($name -ne 'SHA256SUMS.txt' -and $digest -cne $checksums[$name]){throw ('Release checksum mismatch: '+$name)}
 $local[$name]=@{path=$file.FullName;size=$file.Length;digest='sha256:'+$digest}
}
$catalog=Get-Content -LiteralPath "$directory/workspace-update.json" -Raw | ConvertFrom-Json
if($catalog.tag -cne $tag -or $catalog.repository -cne $repo){throw 'Managed metadata does not match the release'}
foreach($entry in $catalog.packages.PSObject.Properties.Value){
 $asset=$local[$entry.asset]
 if(!$asset -or $asset.size -ne $entry.size -or $asset.digest -cne ('sha256:'+$entry.sha256)){throw 'Managed package does not match its asset'}
}
$existing=gh release view $tag --repo $repo --json isDraft,targetCommitish 2>$null
if($LASTEXITCODE){
 gh release create $tag --repo $repo --draft --prerelease --target $source --title "Froststock $tag" --notes 'Independent Froststock release for One Workspace. Install or update it from Workspace Updates.'
 if($LASTEXITCODE){throw 'Froststock draft creation failed'}
}else{
 $draft=$existing | ConvertFrom-Json
 if(!$draft.isDraft -or $draft.targetCommitish -cne $source){throw 'An immutable or mismatched release already uses this tag'}
}
function GetAssets {
 $rows=gh api "repos/$repo/releases?per_page=20" | ConvertFrom-Json
 if($LASTEXITCODE){throw 'Froststock release metadata unavailable'}
 $release=$rows | Where-Object tag_name -CEQ $tag
 if(!$release){throw 'Froststock release missing'}
 return $release.assets
}
function VerifyAsset($asset) {
 $expected=$local[$asset.name]
 if(!$expected -or $asset.state -cne 'uploaded' -or $asset.digest -cne $expected.digest -or $asset.size -ne $expected.size){throw ('Uploaded asset differs: '+$asset.name)}
}
$existingAssets=@(GetAssets)
$names=@{}
foreach($asset in $existingAssets){VerifyAsset $asset; $names[$asset.name]=$true}
$missing=@($local.Keys | Sort-Object | Where-Object {!$names.ContainsKey($_)} | ForEach-Object {$local[$_].path})
if($missing.Count){
 gh release upload $tag --repo $repo @missing
 if($LASTEXITCODE){throw 'Froststock asset upload failed'}
}
$uploaded=@(GetAssets)
if($uploaded.Count -ne $local.Count){throw 'Froststock release is missing assets'}
foreach($asset in $uploaded){VerifyAsset $asset}
gh release edit $tag --repo $repo --draft=false --prerelease
if($LASTEXITCODE){throw 'Froststock publication failed'}
$published=(gh api "repos/$repo/commits/$tag" --jq '.sha').Trim()
if($LASTEXITCODE -or $published -cne $source){throw 'Published Froststock tag differs from the source'}
$url="https://github.com/$repo/releases/tag/$tag"
Write-Output "Published and verified $($uploaded.Count) Froststock assets. $url"
