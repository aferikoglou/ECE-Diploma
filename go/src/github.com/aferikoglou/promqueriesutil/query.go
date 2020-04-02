package promqueriesutil

import (
         "io/ioutil"
         "log"
         "net/http"
         "encoding/json"
	 "strconv"
)

func Query(prometheus_endpoint string, metric string) map[int]float64 /* map[int]int */{
	var endpoint string = prometheus_endpoint + "/api/v1/query?query=" + metric

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

	/***** Get the actual data *****/

        m := f.(map[string]interface{})

        b, err := json.MarshalIndent(m, "", "  ")
        if err != nil {
                log.Fatalln(err)
        }

	var o map[string]interface{}
	json.Unmarshal(b, &o)

	arr := o["data"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})["value"].([]interface{})

        timestamp := int(arr[0].(float64))

	// val, _ := strconv.Atoi(arr[1].(string))
	val, err := strconv.ParseFloat(arr[1].(string), 64)
	if err != nil {
		log.Fatalln(err)
	}

	/***** Create output map  *****/

	// var output_map map[int]int
	// output_map = make(map[int]int)
	var output_map map[int]float64
	output_map = make(map[int]float64)

	output_map[timestamp] = val

	return output_map
}
