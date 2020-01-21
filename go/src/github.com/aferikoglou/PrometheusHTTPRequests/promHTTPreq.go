package main

import (
	"fmt"
	"log"
	"time"
	"os"
	"strconv"
	"github.com/aferikoglou/promqueriesutil"
)

const PROMETHEUS_ENDPOINT string = "http://192.168.1.145:30000"

const STEP int = 1

func main() {
	fmt.Println(" ------------- ")
	fmt.Println("|QUERY RESULTS|")
	fmt.Println(" ------------- ")

	/***** Get command line arguments *****/
	var metric string = os.Args[1]
	var range_query string = os.Args[2] //y or n

	var output []byte

	if range_query == "y" {
		var timestamp int64 = int64(time.Now().Unix())
		var r_str string = os.Args[3]

		r, err := strconv.Atoi(r_str)
		if err != nil {
			log.Fatalln(err)
		}

		output = promqueriesutil.RangeQuery(PROMETHEUS_ENDPOINT, metric, timestamp - int64(r-1),  timestamp, STEP)
	}

	if range_query == "n" {
		output = promqueriesutil.Query(PROMETHEUS_ENDPOINT, metric)
	}

	fmt.Println(string(output))
}
