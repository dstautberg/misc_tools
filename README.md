# Misc Tools

This is just a collection of miscellaneous tools.

## dstat.ps1 - Displays current disk info

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
