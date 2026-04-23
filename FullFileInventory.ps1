<#
.SYNOPSIS
    Scans all active drives and appends file metadata to a central CSV.
    
.DESCRIPTION
    To run this script:
    1. Open PowerShell (as Administrator for full access).
    2. Navigate to the script's folder: cd "C:\path\to\script"
    3. Execute: powershell -File FullFileInventory.ps1
    
    Note: If you get an execution policy error, run:
    Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process
#>

# Get all Fixed and Removable drives that actually have media inserted
$drives = [System.IO.DriveInfo]::GetDrives() | Where-Object { 
    $_.IsReady -and ($_.DriveType -eq 'Fixed' -or $_.DriveType -eq 'Removable') 
}

$outputPath = ".\FullFileInventory.csv"
$timeStamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"

foreach ($drive in $drives) {
    $targetPath = $drive.RootDirectory.FullName
    Write-Host "Scanning Drive: $targetPath..." -ForegroundColor Cyan

    # Recurse through all files, skipping folders where access is denied
    Get-ChildItem -Path $targetPath -Recurse -File -ErrorAction SilentlyContinue | ForEach-Object {
        [PSCustomObject]@{
            DateScanned = $timeStamp
            HostName    = $env:COMPUTERNAME
            Drive       = $targetPath
            FilePath    = $_.DirectoryName
            FileName    = $_.Name
            Size_MB     = [math]::Round($_.Length / 1MB, 2)
            LastMod     = $_.LastWriteTime
        }
    } | Export-Csv -Path $outputPath -NoTypeInformation -Append
}

Write-Host "Success! Data appended to $outputPath" -ForegroundColor Green
