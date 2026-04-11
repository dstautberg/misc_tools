import schedule
import time
from backup_phone import run_backup

def job():
    print("\n" + "="*60)
    print("Starting scheduled backup...")
    print("="*60 + "\n")
    run_backup()

# Schedule the backup to run daily at a certain time
t = "18:05"
schedule.every().day.at(t).do(job)

print(f"Backup scheduler started. Waiting for scheduled time ({t})...")
print("Press Ctrl+C to stop the scheduler.\n")

while True:
    schedule.run_pending()
    time.sleep(60)  # Check every minute