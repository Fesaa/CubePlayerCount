import os
import zipfile
import asyncio
import aiohttp
from io import BytesIO

years = ["2022"]
months = ["9"]
days = [str(i) for i in range(10,18)]

dates = [(str(i), "9") for i in range(10,11)]


async def get_sql(day: str, month: str, year: str, cs: aiohttp.ClientSession) -> None:
    print(f"Retrieving data for {day}-{month}-{year}...")
    async with cs.get(f"https://dl.minetrack.me/Java/{day}-{month}-{year}.sql.zip") as req:
        if req.status != 200:
            return None
        file_name = f"{year}-{month}-{day}.sql"
        zip= zipfile.ZipFile(BytesIO(await req.read()))
        zip.extractall('./sql')
        os.rename(f"./sql/Minetrack/database_copy_{day}-{month}-{year}.sql", f"./sql/{file_name}")

async def main():
    async with aiohttp.ClientSession() as cs:
        for year in years:
            for date in dates:
                await get_sql(date[0], date[1], year, cs)
    os.removedirs("./sql/Minetrack")
    print("Done!")


asyncio.run(main())
