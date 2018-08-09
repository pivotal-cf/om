trap {
  write-error $_
  exit 1
}

$env:GOPATH = Join-Path -Path $PWD "gopath"
$env:PATH = $env:GOPATH + "/bin;C:/go/bin;C:/bin;" + $env:PATH

cd $env:GOPATH/src/github.com/cloudfoundry/bosh-utils

if ((Get-Command "go.exe" -ErrorAction SilentlyContinue) -eq $null -Or (go.exe version) -ne "go version go1.10 windows/amd64")
{
  Write-Host "Installing Go 1.10!"
  Invoke-WebRequest https://storage.googleapis.com/golang/go1.10.windows-amd64.msi -OutFile go.msi

  $p = Start-Process -FilePath "msiexec" -ArgumentList "/passive /norestart /i go.msi" -Wait -PassThru

  if($p.ExitCode -ne 0)
  {
    throw "Golang MSI installation process returned error code: $($p.ExitCode)"
  }
  Write-Host "Go is installed!"
}

if ((Get-Command "tar.exe" -ErrorAction SilentlyContinue) -eq $null)
{
  Write-Host "Installing tar!"
  New-Item -ItemType directory -Path C:\bin -Force

  Invoke-WebRequest https://s3.amazonaws.com/bosh-windows-dependencies/tar-1490035387.exe -OutFile C:\bin\tar.exe

  Write-Host "tar is installed!"
}

go.exe version

go.exe install github.com/cloudfoundry/bosh-utils/vendor/github.com/onsi/ginkgo/ginkgo
ginkgo.exe -r -keepGoing -skipPackage="vendor"
if ($LastExitCode -ne 0)
{
    Write-Error $_
    exit 1
}
