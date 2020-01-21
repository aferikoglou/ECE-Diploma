package promqueriesutil

import (
         "io/ioutil"
         "log"
         "net/http"
         "encoding/json"
)

func Query(prometheus_endpoint string, metric string) []byte {
	var command string = prometheus_endpoint + "/api/v1/query?query=" + metric

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
