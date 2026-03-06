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
- Number of items with duplicate names
- Detailed list of duplicates with ID, name, username, and password
- Any password containing a search string is highlighted in red
