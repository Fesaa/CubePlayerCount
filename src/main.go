package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/montanaflynn/stats"
	"github.com/wcharczuk/go-chart/v2"
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

func main() {
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	start_time := time.Now()
	files, err := IOReadDir("sql/")
	if err != nil {
		log.Fatal(err)
	}

	DataMap := map[string][]Record{}
	var LineCounter int = 0

	for _, file := range files {
		date := strings.Split(file, "-")
		year := date[0]

		db, err := sql.Open("sqlite3", "sql/"+file)
		if err != nil {
			fmt.Println(file)
			log.Fatal(err)
		}
		rows, err := db.Query("SELECT timestamp, playerCount FROM 'pings' WHERE ip = \"play.cubecraft.net\" AND playerCount NOT NULL;")
		if err != nil {
			fmt.Println(file)
			log.Fatal(err)
		}

		var DataList []Record
		var CurrentMinute int = -1
		var CurrentEpoch int64
		var MinuteCounts []int
		for rows.Next() {
			LineCounter++

			var Epoch int64
			var PlayerCount int
			rows.Scan(&Epoch, &PlayerCount)
			CurrentTime := EpochToTime(Epoch)
			if err != nil {
				log.Fatal(err)
			}

			if CurrentMinute != CurrentTime.Minute() {
				CurrentMinute = CurrentTime.Minute()
				CurrentEpoch = customEpoch(CurrentTime.Month(), CurrentTime.Day(), CurrentTime.Hour(), CurrentTime.Day())
				avg := avgOfArray(MinuteCounts)
				if avg > 100 {
					DataList = append(DataList, Record{CurrentEpoch, avg})
				}
				MinuteCounts = []int{}
			} else {
				MinuteCounts = append(MinuteCounts, PlayerCount)
			}

		}
		DataMap[year] = append(DataMap[year], DataList...)
		db.Close()
	}
	start_graph_time := time.Now()

	var chartSeries []chart.Series
	var Annotations []chart.Value2

	var counter int = 1
	var means []float64
	var standDerv []float64

	years := []string{
		"2020",
		"2021",
		"2022",
	}
	for _, year := range years {
		data := DataMap[year]
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
			if CurrentDay == -1 {
				CurrentDay = EpochToTime(record.Epoch).Day()
			}

			XValues = append(XValues, float64(record.Epoch))
			YValues = append(YValues, float64(record.PlayerCount))

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
			Name:    year,
			XValues: XValues,
			YValues: YValues,
			Style: chart.Style{
				StrokeColor: chart.GetDefaultColor(counter).WithAlpha(uint8(100 / 3 * counter)),
				FillColor:   chart.GetDefaultColor(counter).WithAlpha(uint8(100 / 3 * counter)),
			},
		})

		for Epoch, max := range MaxMap {
			Annotations = append(Annotations, chart.Value2{
				XValue: float64(Epoch),
				YValue: float64(max),
				Label:  "MAX " + year + ": " + fmt.Sprintf("%v", max),
				Style: chart.Style{
					FontSize:    6,
					StrokeColor: chart.GetDefaultColor(counter),
				},
			})
		}
		for Epoch, min := range MinMap {
			Annotations = append(Annotations, chart.Value2{
				XValue: float64(Epoch),
				YValue: float64(min),
				Label:  "MIN " + year + ": " + fmt.Sprintf("%v", min),
				Style: chart.Style{
					FontSize:    6,
					StrokeColor: chart.GetDefaultColor(counter),
				},
			})
		}
		counter++
		mean, _ := stats.Mean(YValues)
		mean = math.Floor(mean*100) / 100
		means = append(means, mean)
		statsDer, _ := stats.StandardDeviation(YValues)
		statsDer = math.Floor(statsDer*100) / 100
		standDerv = append(standDerv, statsDer)
	}

	chartSeries = append(chartSeries, chart.ContinuousSeries{
		Name: "Means: ",
		Style: chart.Style{
			StrokeColor: chart.ColorWhite,
		},
	})

	for index, year := range years {
		chartSeries = append(chartSeries, chart.ContinuousSeries{
			Name: year + ": " + fmt.Sprintf("%v", means[index]),
			Style: chart.Style{
				StrokeColor: chart.ColorWhite,
			},
		})
	}
	chartSeries = append(chartSeries, chart.ContinuousSeries{
		Name: "Stand. der: ",
		Style: chart.Style{
			StrokeColor: chart.ColorWhite,
		},
	})

	for index, year := range years {
		chartSeries = append(chartSeries, chart.ContinuousSeries{
			Name: year + ": " + fmt.Sprintf("%v", standDerv[index]),
			Style: chart.Style{
				StrokeColor: chart.ColorWhite,
			},
		})
	}

	chartSeries = append(chartSeries, chart.AnnotationSeries{
		Annotations: Annotations,
	})

	graph := chart.Chart{
		Title:  "Cube's player count",
		Height: 900,
		Width:  2000,
		XAxis: chart.XAxis{
			Name:         "Time",
			TickPosition: chart.TickPositionBetweenTicks,
			ValueFormatter: func(v interface{}) string {
				typed := int64(v.(float64))
				typedDate := EpochToTime(typed)
				return fmt.Sprintf("%d-%d", typedDate.Month(), typedDate.Day())
			},
		},
		YAxis: chart.YAxis{
			Name: "PlayerCount",
		},
		Series: chartSeries,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	f, _ := os.Create("output/output.png")
	defer f.Close()
	graph.Render(chart.PNG, f)

	fmt.Println("Total Run Time: ", time.Now().Sub(start_time))
	fmt.Println("Reading sqlite: ", start_graph_time.Sub(start_time))
	fmt.Println("\tTotal number of lines read: ", LineCounter)
	fmt.Println("Graph Processing Time: ", time.Now().Sub(start_graph_time))
}
