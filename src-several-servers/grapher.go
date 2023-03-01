package main

import (
	"fmt"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func merge(m map[int][]ChannelData) map[string][]Record {
	out := make(map[string][]Record)
	for _, v := range m {
		for _, cd := range v {
			out[cd.ServerIp] = append(out[cd.ServerIp], cd.R)
		}
	}
	return out
}

func createGraph(ChannelDataMap map[int][]ChannelData, start_time time.Time, servers []string) (chart.Chart, time.Time) {
	DataMap := merge(ChannelDataMap)
	start_graph_time := time.Now()
	fmt.Println("SQL done in", start_graph_time.Sub(start_time))

	var chartSeries []chart.Series
	var Annotations []chart.Value2

	var colourMap = make(map[string]drawing.Color)

	for _, serverIp := range servers {
		colourMap[serverIp] = drawing.Color{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255)), B: uint8(rand.Intn(255)), A: 255}
	}

	var backgroundBlack = drawing.Color{R: 219, G: 213, B: 188, A: 255}

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
					FontColor:   drawing.ColorWhite,
					FillColor:   colourMap[serverIp],
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
					FontColor:   drawing.ColorWhite,
					FillColor:   colourMap[serverIp],
					StrokeColor: colourMap[serverIp],
				},
			})
		}
	}

	chartSeries = append(chartSeries, chart.AnnotationSeries{
		Annotations: Annotations,
	})

	graph := chart.Chart{
		Title: "Player counts",
		TitleStyle: chart.Style{
			FillColor: backgroundBlack,
		},
		Height: 1000,
		Width:  2000,
		XAxis: chart.XAxis{
			Name:         "Date",
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
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph, chart.Style{
			FillColor: backgroundBlack,
		}),
	}
	return graph, start_graph_time
}
