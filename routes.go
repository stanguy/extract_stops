package main


import (
	"fmt"
	"github.com/moovweb/gokogiri"
	"github.com/moovweb/gokogiri/xpath"
	"github.com/stanguy/extract_stops/gtfsreader"
	"github.com/stanguy/gomaps/polyline"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const ROUTES_FILENAME = "routes.txt"
const XML_FILENAME = "star_ligne_itineraire.kml"

type Route struct {
	Id     string
	Name   string
	Colors []string

	Stops []string
	Lines []string
}

func readroutesxml(basedir string) map[string][]string {
	filename := basedir + "/" + XML_FILENAME
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := ioutil.ReadAll(file)
	doc, err := gokogiri.ParseXml(data)
	defer doc.Free()

	xp := doc.DocXPathCtx()
	xp.RegisterNamespace("ns", "http://www.opengis.net/kml/2.2")

	folder_expr := xpath.Compile("//ns:Folder/ns:Placemark")

	folders, err := doc.Search(folder_expr)
	if err != nil {
		log.Fatal(err)
	}

	line_expr := xpath.Compile("ns:ExtendedData//ns:SimpleData[@name='LI_NUM']")
	coords_expr := xpath.Compile("*//ns:LineString/ns:coordinates")

	paths := make(map[string][]string)

	for _, folder := range folders {
		line_short_elems, _ := folder.Search(line_expr)
		if len(line_short_elems) == 0 {
			// maybe log it ?
			continue
		}
		line_short := line_short_elems[0].InnerHtml()

		line_coords, name_exists := paths[line_short]
		if !name_exists {
			line_coords = make([]string, 0)
		}

		coords_elems, _ := folder.Search(coords_expr)

		if 0 == len(coords_elems) {
			coords_elems, _ = folder.Search("ns:LineString/ns:coordinates")
		}

		for _, elem := range coords_elems {
			line_coords = append(line_coords, elem.InnerHtml())
		}
		paths[line_short] = line_coords
	}

	return paths
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
	bgcolor := reader.Headers["route_color"]
	fgcolor := reader.Headers["route_text_color"]

	routes := make([]Route,0)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		route := Route{
			Id:     line[route_id],
			Name:   line[short_name],
			Colors: []string{line[fgcolor], line[bgcolor]},
		}
		route_paths, name_exists := paths[line[short_name]]
		if name_exists {
			real_paths := make([]string, len(route_paths))
			for i, line := range route_paths {
				coords_set := make([][]float64, 0)
				for _, coords_str_set := range strings.Split(line, " ") {
					coords := strings.Split(coords_str_set, ",")
					if len(coords) != 3 {
						continue
					}
					coords_set = append(coords_set, []float64{
						atof(coords[1]), atof(coords[0]),
					})
				}
				real_paths[i] = polyline.Encode(coords_set)
			}
			route.Lines = real_paths
		}
		routes = append(routes, route)
	}
	
	return routes
}
