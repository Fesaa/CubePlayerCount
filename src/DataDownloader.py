import os
import zipfile
import requests
from io import BytesIO

years = ["2020", "2021", "2022"]
months = ["9"]
days = [str(i) for i in range(1,8)]

for year in years:
    for month in months:
        for day in days:
            print(f"Retrieving data for {day}-{month}-{year}...")
            url = f"https://dl.minetrack.me/Java/{day}-{month}-{year}.sql.zip"
            file_name = f"{year}-{month}-{day}.sql"
            req = requests.get(url)
            zip= zipfile.ZipFile(BytesIO(req.content))
            zip.extractall('./sql')
            os.rename(f"./sql/Minetrack/database_copy_{day}-{month}-{year}.sql", f"./sql/{file_name}")
os.removedirs("./sql/Minetrack")
print("Done!")