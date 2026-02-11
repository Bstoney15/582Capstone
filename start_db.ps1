# Ensure data directory exists
$dataDir = Join-Path $PWD "data"
if (-not (Test-Path $dataDir)) {
    New-Item -ItemType Directory -Path $dataDir | Out-Null
}

# Run MySQL container
# -d: Run in background
# --name: Name the container 'mysql-container'
# -v: Mount local ./data folder to /var/lib/mysql in the container for persistence
# -e: Set root password (change 'secret' to your desired password)
# -p: Map port 3306 to localhost:3306
docker run -d `
  --name mysql-container `
  -v "${dataDir}:/var/lib/mysql" `
  -e MYSQL_ROOT_PASSWORD=secret `
  -p 3306:3306 `
  mysql:latest

Write-Host "MySQL container started with root password 'secret'"
