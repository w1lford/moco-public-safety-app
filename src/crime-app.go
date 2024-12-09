package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paulmach/orb"
	_ "github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
	_ "github.com/shaxbee/go-spatialite"
)

type Geolocation struct {
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	HumanAddress string `json:"human_address"`
}

type Incident struct {
	IncidentID           string            `json:"incident_id"`
	OffenceCode          string            `json:"offence_code"`
	CaseNumber           string            `json:"case_number"`
	StartDate            string            `json:"start_date"`
	NibrsCode            string            `json:"nibrs_code"`
	Victims              string            `json:"victims"`
	CrimeName1           string            `json:"crimename1"`
	CrimeName2           string            `json:"crimename2"`
	CrimeName3           string            `json:"crimename3"`
	District             string            `json:"district"`
	City                 string            `json:"city"`
	State                string            `json:"state"`
	ZipCode              string            `json:"zip_code"`
	Agency               string            `json:"agency"`
	Place                string            `json:"place"`
	Sector               string            `json:"sector"`
	Beat                 string            `json:"beat"`
	Pra                  string            `json:"pra"`
	AddressStreet        string            `json:"address_street"`
	StreetType           string            `json:"street_type"`
	Latitude             string            `json:"latitude"`
	Longitude            string            `json:"longitude"`
	PoliceDistrictNumber string            `json:"police_district_number"`
	Geolocation          Geolocation       `json:"geolocation"`
	ComputedRegions      map[string]string `json:",omitempty"` // To hold computed region keys and values (e.g., ":@computed_region_vu5j_pcmz")
}

func main() {
	url := "https://data.montgomerycountymd.gov/resource/icn6-v9z3.json"

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	s := []Incident{}
	//fmt.Println(string(body))
	json.Unmarshal(body, &s)

	db, err := sql.Open("spatialite", "../data/crime-app.sqlite")

	rows, err := db.Query("SELECT pk_uid, name, ST_AsText(geometry) FROM montgomery_county_voter_precincts WHERE name IS NOT NULL")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	precinct_polygons := geojson.NewFeatureCollection()

	for rows.Next() {
		var pk_uid string
		var name string
		var geom string
		if err := rows.Scan(&pk_uid, &name, &geom); err != nil {
			log.Fatal(err)
		}
		p, _ := wkt.UnmarshalMultiPolygon(geom)
		f := geojson.NewFeature(p)
		f.Properties["name"] = name
		f.Properties["pk_uid"] = pk_uid
		precinct_polygons.Append(f)
	}

	//do a spatial intersect with the polygons

	//create a map here to store the intersect counts

	var counts map[string]int
	counts = make(map[string]int)

	for _, v := range s {
		lat, _ := strconv.ParseFloat(v.Latitude, 64)
		lon, _ := strconv.ParseFloat(v.Longitude, 64)
		p := orb.Point{lon, lat}

		for _, f := range precinct_polygons.Features {
			if f.Geometry.Bound().Contains(p) {
				fmt.Println("Point inside polygon:", f.Properties["name"])

				counts[f.Properties["pk_uid"].(string)] += 1
			}
		}
	}
	//now that we have the counts, insert or update into database by pk_uid

	for key, val := range counts {
		result, _ := db.Exec(`UPDATE montgomery_county_voter_precincts SET total_crimes = ? WHERE pk_uid = ?`, val, key)
		fmt.Println(result.RowsAffected())
	}

	//fmt.Println(precinct_polygons)

}
