package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// current configuration values to use for this application
// these settings can be set in the Configuration.json file
// in the current working directory
var outputFile string
var pullURL string
var hoursToPull int

// The Configuration is just an array of config items
type Configuration struct {
	Configs []Config `json:"Configuration"`
}

// Config struct is a simple key value pair stored in json
//  Example:
// 	{
// 		"Key": "Sample Key",
// 		"Value": "Sample Value"
// 	}
type Config struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func main() {
	pullConfig()

	// make an http request to the elasticsearch engine
	//resp, httpErr := http.Get("http://localhost:9200/logstash-2015.05.18/_search?pretty=true")
	resp, httpErr := http.Get(pullURL)

	if httpErr != nil {
		fmt.Println(httpErr)
	}
	defer resp.Body.Close()

	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		fmt.Println(readAllErr)
	}

	//writeErr := ioutil.WriteFile("./"+outputFile, body, 0666)
	text := compressText(body, outputFile)
	writeErr := ioutil.WriteFile("./"+outputFile+".gz", text.Bytes(), 0666)

	if writeErr != nil {
		fmt.Println(writeErr)
	}
}

func pullConfig() {
	// first, pull the settings.json file from this directory.
	// if it doesn't exist, panic
	jsonFile, jsonErr := os.Open("Settings.json")

	if jsonErr != nil {
		fmt.Println(jsonErr)
		panic("Can't load settings from Settings.json, quiting")
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// init the config array
	var configs Configuration

	// unmarshal the byteArray into the configuration object
	json.Unmarshal(byteValue, &configs)

	for i := 0; i < len(configs.Configs); i++ {
		var iter = configs.Configs[i]

		switch strings.ToLower(iter.Key) {
		case "hourstopull":
			temp, convErr := strconv.Atoi(iter.Value)
			if convErr != nil {
				fmt.Println("HoursToPull invalid, expected an integer, actual value: " + iter.Value)
			} else {
				hoursToPull = temp
			}
		case "pullurl":
			pullURL = iter.Value
		case "outputfile":
			outputFile = iter.Value
		default:
			fmt.Println(iter.Key + " is not a tracked key")
		}
	}
}

func compressText(text []byte, filename string) (buf bytes.Buffer) {
	//var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	// setting the header fields is optional
	zw.Name = filename + ".gz"

	_, err := zw.Write(text)
	if err != nil {
		log.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		fmt.Println("Error closing gzip")
		log.Fatal(err)
	}
	return
}
