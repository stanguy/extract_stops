package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
)

const STOPS_FILENAME = "stops.txt"
const ROUTES_FILENAME = "routes.txt"
const TRIPS_FILENAME = "trips.txt"
const STOPTIMES_FILENAME = "stop_times.txt"

type ExportedData struct {
	Stops []StopStation
	Routes []Route
}

func main() {
	log.SetFlags(log.Llongfile)

	var basedir = flag.String("b", "", "Base directory for CSV files")
	flag.Parse()
	var stops_file string
	var routes_file string
	var stoptimes_file string
	var trips_file string
	if len(*basedir) > 0 {
		stops_file = *basedir + "/" + STOPS_FILENAME
		routes_file = *basedir + "/" + ROUTES_FILENAME
		trips_file = *basedir + "/" + TRIPS_FILENAME
		stoptimes_file = *basedir + "/" + STOPTIMES_FILENAME
	} else {
		stops_file = STOPS_FILENAME
		routes_file = ROUTES_FILENAME
		trips_file = TRIPS_FILENAME
		stoptimes_file = STOPTIMES_FILENAME
	}

	route_stops := readstoptimes(stoptimes_file,readtrips(trips_file))
	
	data := ExportedData{
		Stops: readstops(stops_file),
		Routes: readroutes(routes_file),
	}

	for idx, route := range data.Routes {
		all_stops := route_stops[route.Id]
		stops := make([]string,0)
		for stop, _ := range all_stops {
			stops = append(stops,stop)
		}
		data.Routes[idx].Stops = stops
	}
	
	//	fmt.Printf("%d bus stops for %d unique names\n", nb_stops, len(stops_by_name))
	js_data, _ := json.Marshal(data)
	fmt.Printf("%s\n", js_data)

}
