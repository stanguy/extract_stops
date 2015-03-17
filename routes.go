package main


import (
	"github.com/stanguy/extract_stops/gtfsreader"
	"fmt"
	"io"
)

const ROUTES_FILENAME = "routes.txt"
type Route struct {
	Id string
	Name string
	Stops []string
}

func readroutes(basedir string) []Route {
	paths := readroutesxml(basedir)
	routes_file := basedir + "/" + ROUTES_FILENAME
	reader := gtfsreader.NewReader(routes_file)
	if reader == nil {
		fmt.Printf("Unable to open routes file %s\n", routes_file)
		return nil
	}
	defer reader.Close()

	route_id := reader.Headers["route_id"]
	short_name := reader.Headers["route_short_name"]

	routes := make([]Route,0)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		route := Route{
			Id: line[route_id],
			Name: line[short_name],
		}
		routes = append(routes,route)
	}
	
	return routes
}
