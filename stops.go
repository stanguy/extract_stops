package main

import (
	"fmt"
	"github.com/paulsmith/gogeos/geos"
	"github.com/pebbe/go-proj-4/proj"
	"github.com/stanguy/extract_stops/gtfsreader"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const WKS84_SRID = 4326
const G_SRID = 900913

const STOPS_FILENAME = "stops.txt"

func atoi(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func atof(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

func convert_to_cart(point *geos.Geometry) (converted *geos.Geometry, err error) {
	wgs84, err := proj.NewProj("+proj=longlat +ellps=WGS84 +datum=WGS84")
	defer wgs84.Close()
	if err != nil {
		return nil, err
	}
	cart, err := proj.NewProj("+proj=utm +zone=30 +ellps=WGS84 +datum=WGS84 +units=m +no_defs ")
	defer cart.Close()
	if err != nil {
		return nil, err
	}

	orig_x, _ := point.X()
	orig_y, _ := point.Y()

	x, y, err := proj.Transform2(wgs84, cart, proj.DegToRad(orig_x), proj.DegToRad(orig_y))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	converted, err = geos.NewPoint(geos.NewCoord(x, y))
	if err != nil {
		return nil, err
	}
	converted.SetSRID(G_SRID)
	return converted, nil
}

func distance(lhs, rhs *geos.Geometry) (d float64, err error) {
	lhs_cart, err := convert_to_cart(lhs)
	if err != nil {
		log.Fatal(err)
		return 0.0, nil
	}
	rhs_cart, err := convert_to_cart(rhs)
	if err != nil {
		log.Fatal(err)
		return 0.0, nil
	}
	d, err = lhs_cart.Distance(rhs_cart)
	if err != nil {
		return 0.0, err
	}
	return d, nil
}

type Stop struct {
	Name   string
	Pos    [2]float64
	Geom   *geos.Geometry
	StopId int
}

type IndividualStop struct {
	Pos [2]float64
	Id  int
}

type StopStation struct {
	Name    string
	Pos     [2]float64
	Members []IndividualStop
}

func readstops(basedir string) []StopStation {
	stops_file := basedir + "/" + STOPS_FILENAME
	reader := gtfsreader.NewReader(stops_file)
	if reader == nil {
		fmt.Printf("Unable to open stops file %s\n", stops_file)
		return nil
	}
	defer reader.Close()
	stop_lat := reader.Headers["stop_lat"]
	stop_lon := reader.Headers["stop_lon"]
	stop_name := reader.Headers["stop_name"]
	stop_id := reader.Headers["stop_id"]

	stops_by_name := make(map[string][]Stop)

	name_cleaner, _ := regexp.Compile("[ -_\\.]")

	nb_stops := 0

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		pos := [2]float64{atof(line[stop_lat]), atof(line[stop_lon])}

		stop_point, err := geos.NewPoint(geos.NewCoord(pos[1], pos[0]))
		if err != nil {
			log.Fatal(err)
			continue
		}
		stop_point.SetSRID(WKS84_SRID)

		stop := Stop{
			line[stop_name],
			pos,
			stop_point,
			atoi(line[stop_id]),
		}

		simple_name := name_cleaner.ReplaceAllString(strings.ToLower(stop.Name), "")

		content, name_exists := stops_by_name[simple_name]
		if !name_exists {
			content = make([]Stop, 0)
		}
		content = append(content, stop)
		stops_by_name[simple_name] = content

		//		fmt.Printf("%+v\n", pos)
		nb_stops++
	}

	stations := make([]StopStation, 0)

	for _, stops := range stops_by_name {
		if len(stops) == 1 {
			continue
		}
		sub_stops := make([]IndividualStop, len(stops))
		points := make([]*geos.Geometry, len(stops))
		for i := 0; i < len(stops); i++ {
			sub_stops[i] = IndividualStop{
				Pos: stops[i].Pos,
				Id:  stops[i].StopId,
			}
			points[i] = stops[i].Geom
			/* Not really necessary for the moment, we assume data is safe enough
			v, err := distance(stops[0].Geom, stops[i].Geom)
			if err != nil {
				log.Fatal(err)
				continue
			}
			if v > 200 {
				fmt.Printf("d:%f (%s)\n", v, stops[0].Name)
			}
			*/
		}
		mpoints, _ := geos.NewCollection(geos.MULTIPOINT, points...)
		center, _ := mpoints.Centroid()
		x, _ := center.X()
		y, _ := center.Y()
		station := StopStation{
			Name:    stops[0].Name,
			Pos:     [2]float64{x, y},
			Members: sub_stops,
		}
		stations = append(stations, station)
	}
	return stations
}
