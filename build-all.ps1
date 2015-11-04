$MyScript = $MyInvocation.MyCommand.Definition
$MyPath = Split-Path $MyScript
$MyProject = $(Get-Item $MyPath).BaseName

$BuildOpts = @{
    "linux" = @("386", "amd64", "arm", "arm64");
    "darwin" = @("386", "amd64");
    "windows" = @("386", "amd64")
}

Push-Location $MyPath
if (!(Test-Path $MyPath\bin)) { New-Item -Path $MyPath\bin -ItemType Directory }

ForEach ($OS in $BuildOpts.GetEnumerator()) {
    ForEach ($Arch in $OS.Value) {
        $env:GOOS = $OS.Key
        $env:GOARCH = $Arch

        if ($env:GOOS -eq "windows") { $Ext = ".exe" } else { $Ext = $null }

        Try {
            Write-Host "Building: $env:GOOS ($env:GOARCH)"
            & cmd /C go build -o "$MyPath\bin\$MyProject-$env:GOOS-$env:GOARCH$Ext"
        }

        Catch {
            Write-Warning ("Build Failure: $env:GOOS ($env:GOARCH)`r`n`t" + $_.Exception.Message)
        }
    }
}

Pop-Location
