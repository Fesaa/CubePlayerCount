package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type Record struct {
	Epoch       int64
	PlayerCount int
}

func IOReadDir(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

func avgOfArray(a []int) int {
	var avg int
	if len(a) != 0 {
		for _, i := range a {
			avg += i
		}
		return avg / len(a)
	}
	return 0
}

func EpochToTime(s int64) time.Time {
	return time.Unix(s/1000, 0)
}

func customEpoch(month time.Month, day int, hour int, minute int) int64 {
	return time.Date(2000, month, day, hour, minute, 0, 0, time.UTC).UnixMilli()
}

func isElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func rowsProcessor(rows *sql.Rows) map[string][]Record {
	ProcessedData := make(map[string][]Record)
	for rows.Next() {
		var Epoch int64
		var ServerIp string
		var PlayerCount int
		rows.Scan(&Epoch, &ServerIp, &PlayerCount)
		if PlayerCount > 0 {
			ProcessedData[ServerIp] = append(ProcessedData[ServerIp], Record{Epoch, PlayerCount})
		}
	}
	return ProcessedData
}

func main() {
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	start_time := time.Now()
	files, err := IOReadDir("sql/")
	if err != nil {
		log.Fatal(err)
	}

	DataMap := map[string][]Record{}

	for _, file := range files {
		db, err := sql.Open("sqlite3", "sql/"+file)
		if err != nil {
			log.Fatal(err)
		}
		rows, err := db.Query("SELECT timestamp, ip, playerCount FROM 'pings' WHERE playerCount NOT NULL AND (ip IN (\"play.cubecraft.net\", \"play.wynncraft.com\", \"play.gommehd.net\"));")
		if err != nil {
			log.Fatal(err)
		}

		Data := rowsProcessor(rows)
		for serverIp, serverData := range Data {
			DataMap[serverIp] = append(DataMap[serverIp], serverData...)
		}
		db.Close()
	}
	start_graph_time := time.Now()

	var chartSeries []chart.Series
	var Annotations []chart.Value2

	colourMap := map[string]drawing.Color{
		"play.cubecraft.net": {R: 53, G: 0, B: 28, A: 255},
		"play.wynncraft.com": {R: 0, G: 85, B: 40, A: 255},
		"play.gommehd.net": {R: 0, G: 85, B: 80, A: 255},
	}

	var backgroundBlack = drawing.Color{R: 16, G: 12, B: 8, A: 255}

	for serverIp := range DataMap {

		data := DataMap[serverIp]
		var XValues []float64
		var YValues []float64

		MaxMap := map[int64]float64{}
		MinMap := map[int64]float64{}
		var max float64 = 0
		var min float64 = 10000
		var MaxEpoch int64
		var MinEpoch int64

		var CurrentDay int = -1

		for index, record := range data {
			XValues = append(XValues, float64(record.Epoch))
			YValues = append(YValues, float64(record.PlayerCount))

			if CurrentDay == -1 {
				CurrentDay = EpochToTime(record.Epoch).Day()
			}

			if float64(record.PlayerCount) > max {
				max = float64(record.PlayerCount)
				MaxEpoch = record.Epoch
			}

			if float64(record.PlayerCount) < min && float64(record.PlayerCount) > 0 {
				min = float64(record.PlayerCount)
				MinEpoch = record.Epoch
			}

			if CurrentDay != EpochToTime(record.Epoch).Day() || index == len(data)-1 {
				MaxMap[MaxEpoch] = max
				MinMap[MinEpoch] = min
				max, min = 0, 10000
				CurrentDay = EpochToTime(record.Epoch).Day()
				
			}
		}
		chartSeries = append(chartSeries, chart.ContinuousSeries{
			Name:    serverIp,
			XValues: XValues,
			YValues: YValues,
			Style: chart.Style{
				StrokeColor: colourMap[serverIp],
			},
		})
		for e, max := range MaxMap {
			Annotations = append(Annotations, chart.Value2{
				XValue: float64(e),
				YValue: float64(max),
				Label:  "MAX: " + fmt.Sprintf("%v", max),
				Style: chart.Style{
					FontSize:    6,
					FontColor: drawing.ColorWhite,
					FillColor: colourMap[serverIp],
					StrokeColor: colourMap[serverIp],
				},
			})
		}
		for e, min := range MinMap {
			Annotations = append(Annotations, chart.Value2{
				XValue: float64(e),
				YValue: float64(min),
				Label:  "MIN:" + fmt.Sprintf("%v", min),
				Style: chart.Style{
					FontSize:    6,
					FontColor: drawing.ColorWhite,
					FillColor: colourMap[serverIp],
					StrokeColor: colourMap[serverIp],
				},
			})
		}
	}

	chartSeries = append(chartSeries, chart.AnnotationSeries{
		Annotations: Annotations,
	})

	graph := chart.Chart{
		Title:  "Player counts",
		TitleStyle: chart.Style{
			FillColor: backgroundBlack,
			FontColor: drawing.ColorWhite,
		},
		Height: 1000,
		Width:  2000,
		XAxis: chart.XAxis{
			Name:         "Time",
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := int64(v.(float64))
				typedDate := EpochToTime(typed)
				return fmt.Sprintf("%d/%d", typedDate.Day(), typedDate.Month())
			},
		},
		YAxis: chart.YAxis{
			Name: "PlayerCount",
		},
		Series: chartSeries,
		Canvas: chart.Style{
			FillColor: backgroundBlack,
		},
		Background: chart.Style{
			FillColor: backgroundBlack,
			FontColor: drawing.ColorWhite,
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph, chart.Style{
			FillColor: backgroundBlack,
			FontColor: drawing.ColorWhite,
		}),

	}
	f, _ := os.Create("output/output_" + fmt.Sprintf("%v", time.Now().Unix()) + ".png")
	defer f.Close()
	graph.Render(chart.PNG, f)

	fmt.Println("Total Run Time: ", time.Now().Sub(start_time))
	fmt.Println("Reading sqlite: ", start_graph_time.Sub(start_time))
	fmt.Println("Graph Processing Time: ", time.Now().Sub(start_graph_time))
}
