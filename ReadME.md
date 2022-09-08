# CubeCrafts player count grapher
Simple script to compare CubeCrafts player count over the years made in go.

___
# How to use
Clone the repo
```
cd <your folder>
git clone https://github.com/Fesaa/CubePLayerCount
cd src
```
## Download the data
```
python3 -m venv venv/
source venv/bin/activate
pip install -r requirements.txt
```
Set which years, months, and days you want to compare by changing the values in `DataDownloader.py` lines `6-8`.
Finally; run the file
```
python3 DataDownloader.py
```
## Run the go script
Install dependencies & run
```
go get github.com/montanaflynn/stats
go get github.com/wcharczuk/go-chart/v2
go run .
```

Your output can be found in src/output!