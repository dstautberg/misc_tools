function Get-DriveReport {
    # Get the drive letter from your current location
    $driveLetter = (Get-Location).Drive.Name

    if ($null -eq $driveLetter) {
        Write-Warning "Current location is not a drive."
        return
    }

    # Gather data from the Partition, Disk, and Volume layers
    Get-Partition -DriveLetter $driveLetter | ForEach-Object {
        $disk = Get-Disk -Number $_.DiskNumber
        $vol  = Get-Volume -DriveLetter $driveLetter
        
        # Create a custom object with the exact labels you want
        [PSCustomObject]@{
            Label        = if($vol.FileSystemLabel){$vol.FileSystemLabel}else{"[No Label]"}
            Model        = $disk.Model
            Capacity     = "$([math]::Round($disk.Size / 1GB, 2)) GB"
            "Free Space" = "$([math]::Round($vol.SizeRemaining / 1GB, 2)) GB"
        } | Format-List
    }
}

# Alias for quick access
Set-Alias dstat Get-DriveReport
# Usage: Just run 'dstat' in your PowerShell terminal to get the drive report for the current location's drive.
