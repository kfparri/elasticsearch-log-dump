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
	"time"
)

// current configuration values to use for this application
// these settings can be set in the Configuration.json file
// in the current working directory
var hoursToPull int
var baseSearchURL string
var sizeParam int
var prettyOutput bool
var outputFileName string
var outputDirectory string

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

	// this is used to test against my local dataset (elastic sample data set)
	timeToPullFrom := time.Date(2015, 5, 20, 12, 0, 0, 0, time.Local)
	//timeToPullFrom := time.Now().Add(time.Hour * -1 * time.Duration(hoursToPull))

	fmt.Print("Pulling values since ")
	fmt.Println(timeToPullFrom)

	dateVal := strconv.Itoa(timeToPullFrom.Year()) + "-" + strconv.Itoa(int(timeToPullFrom.Month())) + "-" + strconv.Itoa(timeToPullFrom.Day())

	// make an http request to the elasticsearch engine
	//resp, httpErr := http.Get("http://localhost:9200/logstash-2015.05.18/_search?pretty=true")
	// here is the date range search criteria: http://localhost:9200/logstash-*/_search?q=@timestamp:>=2015-05-19&from=0&size=100&pretty=true
	//http://localhost:9200/logstash-*/_search?from=0&size=10000&pretty=true
	//resp, httpErr := http.Get(baseSearchURL + "/?q=@timestamp:>=" + dateVal + "&from=0&size=" + strconv.Itoa(sizeParam) + "&pretty=" + strconv.FormatBool(prettyOutput))

	resp := requestElasticData(dateVal, "0", strconv.Itoa(sizeParam), strconv.FormatBool(prettyOutput))

	//fmt.Println(baseSearchURL + "/?q=@timestamp:>=" + dateVal + "&from=0&size=" + strconv.Itoa(sizeParam) + "&pretty=" + strconv.FormatBool(prettyOutput))

	// if httpErr != nil {
	// 	fmt.Println(httpErr)
	// }

	defer resp.Body.Close()

	body, readAllErr := ioutil.ReadAll(resp.Body)
	if readAllErr != nil {
		fmt.Println(readAllErr)
	}

	//writeErr := ioutil.WriteFile("./"+outputFile, body, 0666)
	text := compressText(body, outputFileName)
	writeErr := ioutil.WriteFile("./"+outputFileName+".gz", text.Bytes(), 0666)

	if writeErr != nil {
		fmt.Println(writeErr)
	}
}

func requestElasticData(dateString string, from string, size string, pretty string) (resp *http.Response) {
	resp, httpErr := http.Get(baseSearchURL + "/?q=@timestamp:>=" + dateString + "&from=" + from + "&size=" + size + "&pretty=" + pretty)

	if httpErr != nil {
		fmt.Println(httpErr)
	}

	return
}

// This function will open up the Settings.json file and load the values that this program needs out of it.
//  it ignores keys that are not used and writes a message of those keys
func pullConfig() {
	// first, pull the settings.json file from this directory.
	// if it doesn't exist, panic
	jsonFile, jsonErr := os.Open("Settings.json")

	// the program cannot proceed unless it gets the settings loaded, if we had an error, panic and quit the program.
	if jsonErr != nil {
		fmt.Println(jsonErr)
		panic("Can't load settings from Settings.json, quiting")
	}

	defer jsonFile.Close()

	// read all the settings
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// init the config array
	var configs Configuration

	// unmarshal the byteArray into the configuration object
	json.Unmarshal(byteValue, &configs)

	// loop through the configuration and find the keys that we need to use and assign them to the global variables to be used

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
		case "basesearchurl":
			baseSearchURL = iter.Value
		case "outputfilename":
			outputFileName = iter.Value
		case "outputdirectory":
			outputDirectory = iter.Value
		case "pretty":
			temp, convErr := strconv.ParseBool(iter.Value)
			if convErr != nil {
				fmt.Println("Expected boolean value for Pretty, actually received: " + iter.Value)
			} else {
				prettyOutput = temp
			}
		case "sizeparam":
			temp, convErr := strconv.Atoi(iter.Value)
			if convErr != nil {
				fmt.Println("SizeParam invalid, expected an integer, actual value: " + iter.Value)
			} else {
				sizeParam = temp
			}
		default:
			fmt.Println(iter.Key + " is not a tracked key")
		}
	}
}

func compressText(text []byte, filename string) (buf bytes.Buffer) {
	// create new gzip writer
	zw := gzip.NewWriter(&buf)

	// setting the header fields is optional
	zw.Name = filename + ".gz"

	// write the provided text byte array to the buffer
	_, err := zw.Write(text)
	if err != nil {
		log.Fatal(err)
	}

	// close the gzip writer
	if err := zw.Close(); err != nil {
		fmt.Println("Error closing gzip")
		log.Fatal(err)
	}

	return
}
