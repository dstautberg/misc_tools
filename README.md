# Misc Tools

This is just a collection of miscellaneous tools.

---

## *Powershell Script: dstat.ps1*

A lightweight PowerShell utility to provide a report of the drive you are currently working in.

### Features

* Detects which drive you are currently in (e.g., `C:`, `F:`, etc.) and reports on that specific disk.
* Combines the physical **Model** (e.g., My Book 1230) with your custom **Volume Label**.
* Displays Capacity and Free Space in GB (rounded to two decimal places).

### Example Output

```text
Label        : MY_BOOK_01
Model        : My Book 1230
Capacity     : 7452.04 GB
Free Space   : 1205.40 GB
```

### Setup & Installation

#### 1. Enable Script Execution

By default, Windows blocks the execution of `.ps1` files. To allow this script to run, open PowerShell as Administrator and run:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 2. Add to your Profile (Optional)

To use `dstat` as a command from anywhere without typing the full file path, add a function to your PowerShell profile:

1. Open your profile: `notepad $PROFILE`
1. Add this line:

```powershell
function dstat { & "C:\path\to\your\repo\dstat.ps1" }
```

1. Restart PowerShell.

## Usage

Simply navigate to any drive in your terminal and run the script:

```powershell
cd F:
dstat
```

---

## *Python Script: bw_duplicate_entries.py*

A Python utility to analyze an export file from Bitwarden password manager, and identify entries with duplicate names, with support for highlighting passwords containing specific keywords.

### Features

* Loads a Bitwarden JSON export file specified on the command line.
* Identifies and reports entries with duplicate names.
* Highlights entries whose passwords contain configurable search strings (displayed in red). This is intended to highlight passwords that you know are old, so you can delete those entries. Search strings are specified in a `.env` file.

### Setup & Installation

#### 1. Install Dependencies

```powershell
pip install python-dotenv
```

#### 2. Configure Search Strings (Optional)

Create a `.env` file in the same directory with your search keywords:

```env
SEARCH_STRINGS='password,1234,secret'
```

### Usage

```powershell
python bw_duplicate_entries.py "D:\path\to\export.json" # Specify custom export file
```

### Output

The script displays:

* Number of items with duplicate names
* Detailed list of duplicates with ID, name, username, and password
* Any password containing a search string is highlighted in red

---

## *Go Program: backup.go*

A Go utility to recursively backup files from one directory to another with support for bandwidth rate limiting.

### Features

* Recursively copies all files and subdirectories from source to destination
* Automatically creates destination directory and subdirectories if they don't exist
* Optional bandwidth rate limiting (MB/s)
* Preserves file permissions
* Displays progress during backup
* Reports total bytes transferred and elapsed time

### Setup & Installation

#### 1. Install Go

Download and install Go from [golang.org](https://golang.org/dl)

#### 2. Install Dependencies

```powershell
go get github.com/joho/godotenv
```

#### 3. Configure Backup Settings

Create a `.env` file in the same directory with your backup settings:

```env
BACKUP_SOURCE=D:\MyDocuments
BACKUP_DEST=D:\Backups\MyDocuments
TRANSFER_RATE_MB=50
```

**Configuration Options:**

* `BACKUP_SOURCE` (required): Path to the source directory to backup
* `BACKUP_DEST` (required): Path to the destination directory
* `TRANSFER_RATE_MB` (optional): Maximum transfer rate in MB/s (0 or omitted = unlimited)

### Usage & Building

#### Build the executable

```powershell
go build -o backup.exe backup.go
```

#### Run the backup

```powershell
.\backup.exe
```

The script will read settings from `.env` in the current directory and start the backup process.

### Output

The script displays:

* Backup configuration at start (source, destination, transfer rate limit)
* Real-time progress for each file with:
  * Timestamp (updated every 5 seconds)
  * Filename and file size
  * Current bandwidth in MB/s
* Final statistics with total bytes transferred and elapsed time
* Any warnings for files that couldn't be copied

**Example Output:**

```bash
Backup started:
  Source: D:\source
  Destination: D:\backup
  Transfer Rate Limit: 10 MB/s

[2026-03-10 20:13:42] Copying: file1.jpg (0.30 MB) - 0.52 MB/s
[2026-03-10 20:15:17] Copying: largefile.wmv (53.01 MB) - 0.56 MB/s

Backup completed!
  Total transferred: 50,000 MB
  Time elapsed: 2h15m30s
```

---

## Rename Tools

This repository includes two utilities that standardize PDF statement filenames to the format `PaypalStatement-YYYY-MM.pdf`.

* **Python script:** `rename_files.py`
  * Purpose: Rename PDF statement files matching multiple patterns to `PaypalStatement-YYYY-MM.pdf`.
  * Supported patterns:
    * `statement-Mon-YYYY.pdf` (e.g., `statement-Apr-2024.pdf` → `PaypalStatement-2024-04.pdf`)
    * `Statement_YYYYMM.pdf` (e.g., `Statement_201406.pdf` → `PaypalStatement-2014-06.pdf`)
  * Usage:
    * `python rename_files.py` (rename files in current directory)
    * `python rename_files.py /path/to/dir` (rename in specified directory)
    * `python rename_files.py /path/to/dir --dry-run` (preview changes without renaming)
  * Notes: The script skips files that don't match or where a target name already exists. Use `--dry-run` to preview.

* **Go program:** `cmd/rename` (builds to `rename.exe`)
  * Purpose: Same behavior as the Python script, provided as a compiled binary for convenience.
  * Build:

    ```powershell
    go build -o rename.exe ./cmd/rename
    ```

  * Run examples:

    ```powershell
    ./rename.exe . --dry-run     # scan current directory (Windows: rename.exe . --dry-run)
    ./rename.exe "C:\path\to\file.pdf"   # rename a single file
    go run ./cmd/rename . --dry-run         # run without building
    ```

  * Note: The Go binary prints results similarly and pauses for `Enter` before exiting (useful when launched by double-click).

* **Windows Explorer integration:** `rename.reg`
  * The file `rename.reg` contains a registry entry that adds a right-click context menu item "Rename with Go Tool" which invokes `rename.exe` with the selected file path.
  * To enable the context menu, import the `.reg` file (double-click it or use `regedit`) after updating the `rename.exe` path inside the file to match your build location.

Both tools implement the same filename patterns and provide a `--dry-run` preview mode for safe usage.

---

## Pixel 6a Backup Tool

A Python script to backup your Pixel 6a phone to a local directory using ADB (Android Debug Bridge) and BetterADBSync.

### Overview

This tool automates the process of backing up your Pixel 6a's internal storage to a destination folder. It uses ADB to communicate with your phone and BetterADBSync to efficiently synchronize files while excluding hidden system files.

### Prerequisites

* **Python 3.6+** installed on your system
* **ADB (Android Debug Bridge)** installed and added to your system PATH
* **BetterADBSync** Python package (`adbsync` command)
* **python-dotenv** for configuration management
* **USB Debugging** enabled on your Pixel 6a

### Installing Dependencies

1. Install ADB:
   * **Windows**: Download [Android SDK Platform Tools](https://developer.android.com/studio/releases/platform-tools) and add to PATH
   * **macOS**: `brew install android-platform-tools`
   * **Linux**: `sudo apt-get install adb` (Ubuntu/Debian)

2. Install Python packages:

   ```bash
   pip install BetterADBSync python-dotenv
   ```

### Setup

1. **Enable USB Debugging** on your Pixel 6a:

    * Go to Settings > About phone
    * Tap "Build number" 7 times to enable Developer options
    * Go to Settings > System > Developer options
    * Enable "USB debugging"

1. **Configure the backup destination**:

    * Copy `.env.example` to `.env`:

      ```bash
      cp .env.example .env
      ```

    * Edit `.env` and set your backup destination:

      ```text
      BACKUP_DESTINATION=D:\\downloads\\Backup-Pixel6a
      ```

    * Use Windows path format with double backslashes (`\\`) or forward slashes (`/`)

1. **Connect your phone**:

   * Connect your Pixel 6a via USB
   * Accept the USB debugging authorization prompt on your phone

### Usage

Run the backup script:

```bash
python backup_phone.py
```

The script will:

1. Check if ADB is available and your Pixel 6a is connected
2. List all entries in `/sdcard/`
3. Sync each entry individually to your specified destination folder
4. Exclude hidden system files (files/folders starting with `.`)
5. Display real-time progress for each entry
6. Report a summary with success/failure counts

Press `Ctrl+C` to cancel the backup operation at any time. The script will display a summary of completed backups.

### Environment Variables (.env file)

* **`BACKUP_DESTINATION`**: The local path where backups are saved
  * Default: `D:\downloads\Backup-Pixel6a`
  * Example: `BACKUP_DESTINATION=D:\\MyBackups\\Phone` or `BACKUP_DESTINATION=D:/MyBackups/Phone`

### Script Settings

You can modify the following in `backup_phone.py`:

* **`source`**: The path on your phone to backup (default: `/sdcard/`)
* **`--exclude` pattern**: Modify the exclusion pattern to skip different file types (default: `**/.*`)

### Notes

* The script backs up each top-level entry in `/sdcard/` individually for better progress tracking
* Hidden files and folders (starting with `.`) are excluded by default to avoid backing up system files
* The backup is incremental - BetterADBSync only transfers new or modified files

### License

This is a personal backup tool. Feel free to modify and use as needed.
