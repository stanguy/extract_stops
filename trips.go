package main

import (
	"extract_stops/gtfsreader"
	"fmt"
	"io"
	"log"
)

const TRIPS_FILENAME = "trips.txt"
const STOPTIMES_FILENAME = "stop_times.txt"

// find all stops associated with lines
func readtrips(basedir string) map[string]string {
	trips_file := basedir + "/" + TRIPS_FILENAME
	reader := gtfsreader.NewReader(trips_file)
	if reader == nil {
		fmt.Printf("Unable to open trips file %s\n", trips_file)
		return nil
	}
	defer reader.Close()

	route_id := reader.Headers["route_id"]
	trip_id := reader.Headers["trip_id"]

	routes_by_trip := make(map[string]string)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		route := fix_star_id(line[route_id])
		trip := line[trip_id]

		routes_by_trip[trip] = route
	}

	return routes_by_trip
}

func readstoptimes(basedir string, routes_by_trip map[string]string) map[string]map[int]bool {
	stoptimes_file := basedir + "/" + STOPTIMES_FILENAME
	reader := gtfsreader.NewReader(stoptimes_file)
	if reader == nil {
		fmt.Printf("Unable to open stoptimes file %s\n", stoptimes_file)
		return nil
	}
	defer reader.Close()

	stop_id := reader.Headers["stop_id"]
	trip_id := reader.Headers["trip_id"]

	line_stops := make(map[string]map[int]bool)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		stop := atoi(fix_star_id(line[stop_id]))
		trip := line[trip_id]

		route, name_exists := routes_by_trip[trip]
		if !name_exists {
			log.Fatal("Unknown trip/route")
		}
		route_stops, name_exists := line_stops[route]
		if !name_exists {
			route_stops = make(map[int]bool)
		}
		route_stops[stop] = true
		line_stops[route] = route_stops
	}

	return line_stops
}
