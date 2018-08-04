package main

import (
	"io/ioutil"
	"net/http"
)

func main() {
	resp, err := http.Get("http://localhost:9200/logstash-2015.05.18/_search?pretty=true")
	if err != nil {
		//handle error
	}
	defer resp.Body.Close()

	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		// handle error
	}

	//bs := string(body)

	err3 := ioutil.WriteFile("./logs.json", body, 0666)

	if err3 != nil {
		// handle
	}

	//fmt.Printf(bs)
}
