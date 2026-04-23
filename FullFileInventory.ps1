# powershell -ExecutionPolicy Bypass -File "FullFileInventory.ps1"

# Get all drives that have a drive letter (Fixed, Removable, or Network)
$drives = Get-PSDrive -PSProvider FileSystem | Where-Object { $_.Free -ne $null }

$report = foreach ($drive in $drives) {
    # Define the root path (e.g., "C:\")
    $rootPath = $drive.Root
    Write-Host "Scanning $rootPath..." -ForegroundColor Cyan

    try {
        # Recurse through all files
        Get-ChildItem -Path $rootPath -Recurse -File -ErrorAction SilentlyContinue | 
        Select-Object @{Name="Drive"; Expression={$rootPath}},
                      @{Name="Directory"; Expression={$_.DirectoryName}},
                      @{Name="FileName"; Expression={$_.Name}},
                      @{Name="Size(MB)"; Expression={[Math]::Round($_.Length / 1MB, 2)}},
                      LastWriteTime
    }
    catch {
        Write-Warning "Could not fully scan $rootPath"
    }
}

# Output to the console and save to a CSV
$report | Out-GridView  # Opens a searchable window
$report | Export-Csv -Path "FullFileInventory.csv" -NoTypeInformation

Write-Host "Inventory complete. CSV saved." -ForegroundColor Green
