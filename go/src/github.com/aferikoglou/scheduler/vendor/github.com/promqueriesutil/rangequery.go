package promqueriesutil

import (
	 "fmt"
	 "io/ioutil"
         "log"
         "net/http"
         "encoding/json"
	 "strconv"
)

func RangeQuery(prometheus_endpoint string, metric string, start_timestamp int64, end_timestamp int64, step int) map[int]float64 /* map[int]int */ {
	var endpoint string = prometheus_endpoint + "/api/v1/query_range?query=" + metric + "&start=" + fmt.Sprint(int32(start_timestamp)) + "&end=" + fmt.Sprint(int32(end_timestamp)) + "&step=" + fmt.Sprint(step) + "s"

	/***** Execute GET request *****/

	resp, err := http.Get(endpoint)
        if err != nil {
		log.Fatalln(err)
	}

        defer resp.Body.Close()

	/***** Get response body *****/

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                log.Fatalln(err)
        }

        /***** Decode JSON data *****/

        var f interface{}
        er := json.Unmarshal(body, &f)
        if er != nil {
                log.Fatalln(er)
        }

	/***** Get the actual data  *****/

	m := f.(map[string]interface{})

        b, err := json.MarshalIndent(m, "", "  ")
        if err != nil {
                log.Fatalln(err)
        }

	var o map[string]interface{}
        json.Unmarshal(b, &o)

	val_arr := o["data"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})["values"].([]interface{})

	/***** Create output map *****/

	// var output_map map[int]int
	// output_map = make(map[int]int)
	var output_map map[int]float64
	output_map = make(map[int]float64)

	for _ , el := range val_arr {
		timestamp := int(el.([]interface{})[0].(float64))

		// val, _ := strconv.Atoi(el.([]interface{})[1].(string))
		val, err := strconv.ParseFloat(el.([]interface{})[1].(string), 64)
		if err != nil {
			log.Fatalln(err)
		}

		output_map[timestamp] = val
	}

	return output_map
}
