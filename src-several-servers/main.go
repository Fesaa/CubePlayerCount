package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wcharczuk/go-chart/v2"
)

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

func EpochToTime(s int64) time.Time {
	return time.Unix(s/1000, 0)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Need at least one server ip to create a graph from!")
	}
	var servers []string = os.Args[1:]

	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	start_time := time.Now()
	files, err := IOReadDir("sql/")
	if err != nil {
		log.Fatal(err)
	}

	ChannelDataMap := processData(files, servers)
	graph, start_graph_time := createGraph(ChannelDataMap, start_time, servers)

	f, _ := os.Create("output/output_" + fmt.Sprintf("%v", time.Now().Unix()) + ".png")
	defer f.Close()
	e := graph.Render(chart.PNG, f)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println("Total Run Time: ", time.Now().Sub(start_time))
	fmt.Println("Reading sqlite: ", start_graph_time.Sub(start_time))
	fmt.Println("Graph Processing Time: ", time.Now().Sub(start_graph_time))
}
