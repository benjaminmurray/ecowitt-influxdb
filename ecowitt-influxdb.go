package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/martinlindhe/unit"
	"github.com/pelletier/go-toml"
)

type configuration struct {
	Port     int
	Logfile  string
	Names    map[string]string
	Influxdb influxdbConfiguration
}

type influxdbConfiguration struct {
	Address     string
	Token       string
	Org         string
	Bucket      string
	Measurement string
}

var config configuration

func insertData(timestamp time.Time, fields map[string]interface{}) {
	// create new client with default option for server url authenticate by token
	client := influxdb2.NewClient(config.Influxdb.Address, config.Influxdb.Token)
	defer client.Close()
	// Use blocking write client for writes to desired bucket
	writeAPI := client.WriteAPIBlocking(config.Influxdb.Org, config.Influxdb.Bucket)
	// Create point
	p := influxdb2.NewPoint(config.Influxdb.Measurement, nil, fields, timestamp)
	// write point immediately
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		log.Println(fmt.Errorf("writing to database: %w", err))
	}
}

func reportData(w http.ResponseWriter, r *http.Request) {
	originalData := map[string]string{}
	r.ParseForm()
	for k, v := range r.Form {
		originalData[k] = v[0]
	}
	if config.Logfile != "" {
		writeJSON(config.Logfile, originalData)
	}
	timestamp, fields := convertData(originalData)
	fields = renameFields(fields, config.Names)
	insertData(timestamp, fields)
}

func renameFields(oldFields map[string]interface{}, names map[string]string) map[string]interface{} {
	newFields := map[string]interface{}{}
	for oldName, value := range oldFields {
		newName, exists := names[oldName]
		if exists {
			newFields[newName] = value
		} else {
			newFields[oldName] = value
		}
	}
	return newFields
}

func convertData(originalData map[string]string) (time.Time, map[string]interface{}) {
	var timestamp time.Time
	v, exists := originalData["dateutc"]
	if exists {
		var err error
		timestamp, err = convertTime(v)
		if err != nil {
			log.Println(fmt.Errorf("converting field \"dateutc\": %w", err))
		}
	}

	fields := map[string]interface{}{}
	for k, v := range originalData {
		var err error
		switch k {
		case "baromabsin":
			fields["barometer_abs"], err = convertBarometer(v)
		case "baromrelin":
			fields["barometer_rel"], err = convertBarometer(v)

		case "windspeedmph":
			fields["wind_speed"], err = convertWindSpeed(v)
		case "windgustmph":
			fields["wind_gust"], err = convertWindSpeed(v)
		case "maxdailygust":
			fields["wind_gust_daily_max"], err = convertWindSpeed(v)
		case "winddir":
			fields["wind_dir"], err = convertFloat(v)

		case "dailyrainin":
			fields["rain_daily"], err = convertRain(v)
		case "eventrainin":
			fields["rain_event"], err = convertRain(v)
		case "hourlyrainin":
			fields["rain_hourly"], err = convertRain(v)
		case "monthlyrainin":
			fields["rain_monthly"], err = convertRain(v)
		case "rainratein":
			fields["rain_rate"], err = convertRain(v)
		case "totalrainin":
			fields["rain_total"], err = convertRain(v)
		case "weeklyrainin":
			fields["rain_weekly"], err = convertRain(v)
		case "yearlyrainin":
			fields["rain_yearly"], err = convertRain(v)

		case "tempf":
			fields["temperature_out"], err = convertTemperature(v)
		case "tempinf":
			fields["temperature_in_0"], err = convertTemperature(v)
		case "temp1f":
			fields["temperature_in_1"], err = convertTemperature(v)
		case "temp2f":
			fields["temperature_in_2"], err = convertTemperature(v)
		case "temp3f":
			fields["temperature_in_3"], err = convertTemperature(v)
		case "temp4f":
			fields["temperature_in_4"], err = convertTemperature(v)
		case "temp5f":
			fields["temperature_in_5"], err = convertTemperature(v)
		case "temp6f":
			fields["temperature_in_6"], err = convertTemperature(v)
		case "temp7f":
			fields["temperature_in_7"], err = convertTemperature(v)
		case "temp8f":
			fields["temperature_in_8"], err = convertTemperature(v)

		case "solarradiation":
			fields["radiation"], err = convertFloat(v)
		case "uv":
			fields["uv"], err = convertInt(v)

		case "humidity":
			fields["humidity_out"], err = convertFloat(v)
		case "humidityin":
			fields["humidity_in_0"], err = convertFloat(v)
		case "humidity1":
			fields["humidity_in_1"], err = convertFloat(v)
		case "humidity2":
			fields["humidity_in_2"], err = convertFloat(v)
		case "humidity3":
			fields["humidity_in_3"], err = convertFloat(v)
		case "humidity4":
			fields["humidity_in_4"], err = convertFloat(v)
		case "humidity5":
			fields["humidity_in_5"], err = convertFloat(v)
		case "humidity6":
			fields["humidity_in_6"], err = convertFloat(v)
		case "humidity7":
			fields["humidity_in_7"], err = convertFloat(v)
		case "humidity8":
			fields["humidity_in_8"], err = convertFloat(v)

		case "soilmoisture1":
			fields["soil_moisture_1"], err = convertFloat(v)
		case "soilmoisture2":
			fields["soil_moisture_2"], err = convertFloat(v)
		case "soilmoisture3":
			fields["soil_moisture_3"], err = convertFloat(v)
		case "soilmoisture4":
			fields["soil_moisture_4"], err = convertFloat(v)
		case "soilmoisture5":
			fields["soil_moisture_5"], err = convertFloat(v)
		case "soilmoisture6":
			fields["soil_moisture_6"], err = convertFloat(v)
		case "soilmoisture7":
			fields["soil_moisture_7"], err = convertFloat(v)
		case "soilmoisture8":
			fields["soil_moisture_8"], err = convertFloat(v)

		case "wh65batt":
			fields["battery_wh65"], err = convertFloat(v)
		case "batt1":
			fields["battery_in_1"], err = convertFloat(v)
		case "batt2":
			fields["battery_in_2"], err = convertFloat(v)
		case "soilbatt1":
			fields["battery_soil_1"], err = convertFloat(v)
		case "soilbatt2":
			fields["battery_soil_2"], err = convertFloat(v)
		}
		if err != nil {
			log.Println(fmt.Errorf("converting field \"%v\": %w", k, err))
		}
	}
	return timestamp, fields
}

func convertFloat(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	return x, err
}

func convertInt(s string) (int64, error) {
	x, err := strconv.ParseInt(s, 10, 64)
	return x, err
}

func convertTime(s string) (time.Time, error) {
	timestamp, err := time.Parse("2006-01-02 15:04:05", s)
	return timestamp, err
}

// Convert mph to kph
func convertWindSpeed(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	x = (unit.Speed(x) * unit.MilesPerHour).KilometersPerHour()
	x = math.Round(x*10) / 10
	return x, err
}

// Convert inHg to hPa
func convertBarometer(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	x = x / 0.029530
	x = math.Round(x*10) / 10
	return x, err
}

// Convert Fahrenheit to Celsius
func convertTemperature(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	x = unit.FromFahrenheit(x).Celsius()
	x = math.Round(x*10) / 10
	return x, err
}

// Convert inches to millimeters
func convertRain(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	x = (unit.Length(x) * unit.Inch).Millimeters()
	x = math.Round(x*10) / 10
	return x, err
}

func readJSON(fileName string) map[string]string {
	data := map[string]string{}
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Panic(err)
	}
	return data
}

func writeJSON(fileName string, data interface{}) {
	j, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println(err)
		return
	}
	err = ioutil.WriteFile(fileName, j, 0644)
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	// Parse command line arguments
	configFileName := flag.String("config", "ecowitt-influxdb.conf", "Configuration file name")
	testFileName := flag.String("test", "", "Test file name")
	flag.Parse()

	// Read configuration file
	file, err := ioutil.ReadFile(*configFileName)
	if err != nil {
		log.Fatal(fmt.Errorf("reading configuration file \"%v\": %w", configFileName, err))
	}
	config = configuration{}
	err = toml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(fmt.Errorf("parsing configuration file \"%v\": %w", configFileName, err))
	}

	// Do a test run
	if *testFileName != "" {
		data := readJSON(*testFileName)
		timestamp, fields := convertData(data)
		fields = renameFields(fields, config.Names)
		fmt.Println("timestamp =", timestamp)
		json, err := json.MarshalIndent(fields, "", "  ")
		if err == nil {
			fmt.Println("fields =", string(json))
		}
		insertData(timestamp, fields)
		return
	}

	// Start http server
	http.HandleFunc("/data/report/", reportData)
	addr := ":" + strconv.Itoa(config.Port)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("starting server on address \"%v\": %w", addr, err))
	}
}
