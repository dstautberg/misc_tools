import subprocess
import sys
import os
import shutil
import time
import winsound
from dotenv import load_dotenv
from datetime import datetime

def check_device_connected():
    """
    Check if an Android device is connected via ADB.
    If not connected, prompts user to connect and retry.
    Returns True if device is connected, False otherwise.
    """
    check_device = subprocess.run(["adb", "get-state"], capture_output=True, text=True)
    
    if "device" not in check_device.stdout:
        print("\n" + "!"*60)
        print("ERROR: No Pixel detected")
        print("!"*60)
        print("\nPlease connect your phone and enable USB debugging.")
        
        # Play Windows alert sound (more reliable methods)
        try:
            # Try to play the system default beep sound
            winsound.PlaySound("SystemExclamation", winsound.SND_ALIAS)
        except:
            # Fallback to simple beep if that fails
            try:
                winsound.Beep(1000, 500)  # 1000 Hz for 500ms
            except:
                # If all else fails, use MessageBeep
                winsound.MessageBeep(winsound.MB_ICONEXCLAMATION)
        
        print("\nPress SPACE to retry after connecting your phone, or press Ctrl+C to cancel...")
        
        # Wait for spacebar press
        while True:
            if sys.platform == 'win32':
                import msvcrt
                if msvcrt.kbhit():
                    key = msvcrt.getch()
                    if key == b' ':  # Spacebar
                        break
                    elif key == b'\x03':  # Ctrl+C
                        print("\nBackup cancelled.")
                        return False
            time.sleep(0.1)
        
        # Retry device check
        print("\nRetrying device detection...")
        check_device = subprocess.run(["adb", "get-state"], capture_output=True, text=True)
        
        if "device" not in check_device.stdout:
            print("\n✗ Still no device detected. Please check:")
            print("  - USB cable is properly connected")
            print("  - USB debugging is enabled on your phone")
            print("  - You've authorized this computer on your phone")
            print("\nRun the script again after resolving these issues.")
            return False
        else:
            print("✓ Device detected! Continuing with backup...\n")
    
    return True

def get_phone_storage_info():
    """
    Get storage information from the phone.
    Returns a tuple of (total_gb, used_gb, free_gb) or (None, None, None) if failed.
    """
    try:
        # Use 'df' command to get storage info for /sdcard
        result = subprocess.run(
            ["adb", "shell", "df", "/sdcard"],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode == 0:
            # Parse the output (format varies, but typically last line has the data)
            lines = result.stdout.strip().split('\n')
            if len(lines) >= 2:
                # Usually second line has the actual data
                data_line = lines[-1].split()
                if len(data_line) >= 4:
                    # Convert from KB to GB (df typically shows in KB or blocks)
                    # Format is usually: Filesystem  Size  Used  Avail  Use%  Mounted
                    # We need to handle different formats
                    try:
                        # Try to parse the size values (could be in K, M, G format)
                        def parse_size(size_str):
                            """Convert size string (like '10G', '512M') to GB"""
                            size_str = size_str.strip()
                            if size_str.endswith('G'):
                                return float(size_str[:-1])
                            elif size_str.endswith('M'):
                                return float(size_str[:-1]) / 1024
                            elif size_str.endswith('K'):
                                return float(size_str[:-1]) / (1024 * 1024)
                            else:
                                # Assume it's in KB (1K blocks)
                                return float(size_str) / (1024 * 1024)
                        
                        total = parse_size(data_line[1])
                        used = parse_size(data_line[2])
                        free = parse_size(data_line[3])
                        
                        return (total, used, free)
                    except (ValueError, IndexError):
                        pass
        
        return (None, None, None)
    except Exception as e:
        print(f"Warning: Could not retrieve storage info: {e}")
        return (None, None, None)

def run_backup():
    # Start timing
    start_time = datetime.now()
    
    # Load environment variables from .env file
    load_dotenv()
    
    destinations = []
    i = 1
    while True:
        dest = os.getenv(f"BACKUP_DESTINATION_{i}")
        if not dest:
            break
        try:
            os.makedirs(dest, exist_ok=True)
            destinations.append(dest)
            print(f"Created backup destination '{dest}'.")
        except Exception as e:
            print(f"Error creating backup destination '{dest}': {e}")
        i += 1
        
    # Source uses Unix-style path for ADB compatibility
    source = "/sdcard/"
    
    # 2. Check if ADB is available
    adb_path = shutil.which("adb")
    if not adb_path:
        print("Error: ADB not found in PATH.")
        print("Please install Android Platform Tools from:")
        print("https://developer.android.com/studio/releases/platform-tools")
        return
    print(f"Using ADB from: {adb_path}")

    # 3. Check if device is connected via ADB
    if not check_device_connected():
        return

    # 4. List contents of /sdcard/
    print(f"\nListing contents of {source}...")
    list_result = subprocess.run(["adb", "shell", "ls", "/sdcard/"], capture_output=True, text=True)
    
    if list_result.returncode != 0:
        print(f"Error listing directory: {list_result.stderr}")
        return
    
    # Parse the directory listing
    entries = [entry.strip() for entry in list_result.stdout.strip().split('\n') if entry.strip()]
    
    if not entries:
        print("No entries found in /sdcard/")
        return
    
    print(f"Found {len(entries)} entries to backup:")
    for entry in entries:
        print(f"  - {entry}")
    
    success, failed, skipped = [], [], []

    # Loop through each entry and sync it
    for entry in entries:
        if entry in ["Android"]:
            print(f"Skipping '{entry}'")
            skipped.append(entry)
            continue

        for destination in destinations:
            entry_source = f"/sdcard/{entry}"
            entry_dest = destination
            print(f"\n>>> Syncing: {entry} to {entry_dest}")
            
            # Construct the adbsync command with specific folder exclusions
            cmd = [
                "adbsync",
                "--exclude", "**/.*",              # Skip hidden files
                "--exclude", "Android/data/**",    # Recursive skip of private app data
                "--exclude", "Android/obb/**",     # Recursive skip of large game blobs
                "pull",
                entry_source, 
                entry_dest
            ]            

            try:
                # Execute the sync for this entry
                process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
                
                for line in process.stdout:
                    print(line, end="")
                    
                process.wait()
                
                if process.returncode == 0:
                    print(f"Successfully synced: {entry}")
                    if entry not in success: success.append(entry)
                else:
                    msg = f"Failed to sync: {entry} to {entry_dest} (Exit code: {process.returncode})"
                    print(msg)
                    failed.append(msg)
                    
            except KeyboardInterrupt:
                print("\n\nBackup cancelled by user.")
                elapsed_time = datetime.now() - start_time
                minutes, seconds = divmod(int(elapsed_time), 60)
                print(f"\nSummary: {len(success)} succeeded, {len(failed)} failed, {len(skipped)} not processed")
                print(f"Time elapsed: {minutes} minutes, {seconds} seconds")
                return
            except Exception as e:
                print(f"✗ Error syncing {entry}: {e}")
                failed_count += 1
    
    # Calculate elapsed time
    elapsed_time = datetime.now() - start_time
    
    total_seconds = int(elapsed_time.total_seconds())
    hours, remainder = divmod(total_seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    
    # Get phone storage information
    total_gb, used_gb, free_gb = get_phone_storage_info()
    
    # 6. Print summary
    summary = "\n" + "="*60 + "\n"
    summary += f"Backup locations: {destinations}\n"
    summary += f"Backup Started:   {start_time.strftime("%Y-%m-%d %H:%M:%S")}\n"
    summary += f"Elapsed Time:     {hours:02}:{minutes:02}:{seconds:02}\n"
    summary += f"Total entries:    {len(entries)}\n"
    summary += f"Skipped:          {skipped}\n"
    summary += f"Synced:           {success}\n"
    summary += f"Failed:           {failed}\n"
    
    # Display phone storage info
    if total_gb is not None:
        summary += f"Phone Storage:\n"
        summary += f"  Total: {total_gb:.2f} GB\n"
        summary += f"  Used:  {used_gb:.2f} GB ({(used_gb/total_gb*100):.1f}%)\n"
        summary += f"  Free:  {free_gb:.2f} GB ({(free_gb/total_gb*100):.1f}%)\n"
    
    summary += "="*60 + "\n"
    print(summary)
    open(f"{sys.argv[0]}.log", 'a+').write(summary)

if __name__ == "__main__":
    run_backup()
