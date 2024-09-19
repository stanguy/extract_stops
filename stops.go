package main

import (
	"extract_stops/gtfsreader"
	"fmt"
	"github.com/paulsmith/gogeos/geos"
	"github.com/pebbe/proj/v5"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const WKS84_SRID = 4326
const G_SRID = 900913
const MAX_STOP_DISTANCE = 320

const STOPS_FILENAME = "stops.txt"

func atoi(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func atof(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}

type Converter struct {
	ctx      *proj.Context
	pipeline *proj.PJ
	cache    map[*geos.Geometry]*geos.Geometry
}

func NewConverter() *Converter {
	conv := &Converter{}

	conv.ctx = proj.NewContext()

	var err error
	conv.pipeline, err = conv.ctx.Create(`
        +proj=pipeline
        +step +proj=longlat +ellps=WGS84 +datum=WGS84
        +step +proj=utm +zone=30 +ellps=WGS84 +datum=WGS84 +units=m +no_defs
    `)
	if err != nil {
		return nil
	}
	conv.cache = make(map[*geos.Geometry]*geos.Geometry)
	return conv
}

func (self *Converter) to_cart(point *geos.Geometry) (converted *geos.Geometry, err error) {
	if converted, ok := self.cache[point]; ok {
		return converted, nil
	}
	orig_x, _ := point.X()
	orig_y, _ := point.Y()

	x, y, _, _, err := self.pipeline.Trans(proj.Fwd, proj.DegToRad(orig_x), proj.DegToRad(orig_y), 0, 0)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	converted, err = geos.NewPoint(geos.NewCoord(x, y))
	if err != nil {
		return nil, err
	}
	converted.SetSRID(G_SRID)
	self.cache[point] = converted
	return converted, nil
}

func (self *Converter) close() {
	self.pipeline.Close()
	self.ctx.Close()
}

func (self *Converter) distance(lhs, rhs *geos.Geometry) (d float64, err error) {
	lhs_cart, err := self.to_cart(lhs)
	if err != nil {
		log.Fatal(err)
		return 0.0, nil
	}
	rhs_cart, err := self.to_cart(rhs)
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

// temporary structure for computations
type Stop struct {
	Name   string
	Pos    [2]float64
	Geom   *geos.Geometry
	StopId int
}

// sub-component of the full station
type IndividualStop struct {
	Pos [2]float64
	Id  int
}

// this is the exported structure for each stop
type StopStation struct {
	Name    string
	Pos     [2]float64
	Members []IndividualStop
}

func readstops(basedir string, c chan StopStation) {
	stops_file := basedir + "/" + STOPS_FILENAME
	reader := gtfsreader.NewReader(stops_file)
	if reader == nil {
		fmt.Printf("Unable to open stops file %s\n", stops_file)
		return
	}
	defer reader.Close()
	stop_lat := reader.Headers["stop_lat"]
	stop_lon := reader.Headers["stop_lon"]
	stop_name := reader.Headers["stop_name"]
	stop_id := reader.Headers["stop_id"]

	stops_by_name := make(map[string][]Stop)

	name_cleaner, _ := regexp.Compile(`[ -_\\.]`)

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
			atoi(fix_star_id(line[stop_id])),
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

	converter := NewConverter()
	defer converter.close()
	for _, stops := range stops_by_name {
		sorted_stops := make([][]Stop, 0)
		found := false
		for i := 0; i < len(stops); i++ {
			current_stop := stops[i]
			for j, v := range sorted_stops {
				dist, err := converter.distance(current_stop.Geom, v[0].Geom)
				if err != nil {
					log.Fatal(err)
					continue
				}
				if dist < MAX_STOP_DISTANCE {
					// add
					found = true
					sorted_stops[j] = append(sorted_stops[j], current_stop)
					break
				} else if dist >= MAX_STOP_DISTANCE {
					log.Printf("Found stop more distant named %s at %fm", current_stop.Name, dist)
				}
			}
			if !found {
				new_stop := make([]Stop, 1)
				new_stop[0] = current_stop
				sorted_stops = append(sorted_stops, new_stop)
			}
		}
		for _, v := range sorted_stops {
			points := make([]*geos.Geometry, len(v))
			export_stops := make([]IndividualStop, len(v))
			for j, stop := range v {
				points[j] = stop.Geom
				export_stops[j] = IndividualStop{stop.Pos, stop.StopId}
			}
			mpoints, _ := geos.NewCollection(geos.MULTIPOINT, points...)
			center, _ := mpoints.Centroid()
			x, _ := center.X()
			y, _ := center.Y()
			c <- StopStation{
				Name:    stops[0].Name,
				Pos:     [2]float64{x, y},
				Members: export_stops,
			}
		}
	}
	close(c)
}
