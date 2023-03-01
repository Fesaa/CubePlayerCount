package main

import (
	"database/sql"
	"log"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Record struct {
	Epoch       int64
	PlayerCount int
}

type ChannelData struct {
	ServerIp string
	R        Record
}

func channelProcess(m map[int][]ChannelData, ch chan ChannelData) {
	for v := range ch {
		t := EpochToTime(v.R.Epoch).Day()
		m[t] = append(m[t], v)
	}
}

func rowsProcessor(rows *sql.Rows, ch chan ChannelData, wg *sync.WaitGroup) {
	defer wg.Done()
	for rows.Next() {
		var Epoch int64
		var ServerIp string
		var PlayerCount int
		rows.Scan(&Epoch, &ServerIp, &PlayerCount)
		if PlayerCount > 0 {
			ch <- ChannelData{
				ServerIp: ServerIp,
				R:        Record{Epoch, PlayerCount},
			}
		}
	}
}

func processData(files []string, servers []string) map[int][]ChannelData {
	var sqlUseableServers []string
	for _, serverIp := range servers {
		sqlUseableServers = append(sqlUseableServers, "\"" + serverIp + "\"")
	}

	ChannelDataMap := make(map[int][]ChannelData)
	var ch chan ChannelData = make(chan ChannelData)
	wg := sync.WaitGroup{}
	wg.Add(len(files))

	for _, file := range files {
		db, err := sql.Open("sqlite3", "sql/"+file)
		if err != nil {
			log.Fatal(err)
		}
		rows, err := db.Query("SELECT timestamp, ip, playerCount FROM 'pings' WHERE playerCount NOT NULL AND (ip IN (" + strings.Join(sqlUseableServers, ", ") + "));")
		if err != nil {
			log.Fatal(err)
		}

		go rowsProcessor(rows, ch, &wg)
		db.Close()
	}
	go channelProcess(ChannelDataMap, ch)
	wg.Wait()
	close(ch)
	return ChannelDataMap
}
