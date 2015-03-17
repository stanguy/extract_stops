package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
)


type ExportedData struct {
	Stops []StopStation
	Routes []Route
}

func main() {
	log.SetFlags(log.Llongfile)

	var basedir_tmp = flag.String("b", "", "Base directory for CSV files")
	flag.Parse()
	var basedir string
	if len(*basedir_tmp) > 0 {
		basedir = *basedir_tmp
	} else {
		basedir = "."
	}

	data := ExportedData{
		Stops:  readstops(basedir),
		Routes: readroutes(basedir),
	}

	route_stops := readstoptimes(basedir, readtrips(basedir))

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
