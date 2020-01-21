package promqueriesutil

import (
	 "fmt"
	 "io/ioutil"
         "log"
         "net/http"
         "encoding/json"
)

func RangeQuery(prometheus_endpoint string, metric string, start_timestamp int64, end_timestamp int64, step int) []byte {
	var command string = prometheus_endpoint + "/api/v1/query_range?query=" + metric + "&start=" + fmt.Sprint(int32(start_timestamp)) + "&end=" + fmt.Sprint(int32(end_timestamp)) + "&step=" + fmt.Sprint(step) + "s"

	/***** Execute GET request *****/
	resp, err := http.Get(command)
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

        m := f.(map[string]interface{})

        /***** Beautify *****/
        b, err := json.MarshalIndent(m, "", "  ")
        if err != nil {
                log.Fatalln(err)
        }

	return b
}
