package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type StopTime struct {
	TripID    string
	StopID    string
	Arrival   string
	Departure string
}

type Trip struct {
	TripID    string
	RouteID   string
	ServiceID string
}

type TripResponse struct {
	TripID    string             `json:"trip_id"`
	ServiceID string             `json:"service_id"`
	RouteID   string             `json:"route_id"`
	Schedules []ScheduleResponse `json:"schedules"`
}

type ScheduleResponse struct {
	StopID    string `json:"stop_id"`
	Arrival   string `json:"arrival_time"`
	Departure string `json:"departure_time"`
}

func buildTripResponse(
	route string,
	stopTimes []StopTime,
	stopTimesIxByTrip map[string][]int,
	trips []Trip,
	tripsIxByRoute map[string][]int,
) []TripResponse {
	tripIxs, ok := tripsIxByRoute[route]

	if ok {
		resp := make([]TripResponse, 0, len(tripIxs))
		for tripIx := range tripIxs {
			trip := trips[tripIx]
			tripResponse := TripResponse{
				TripID:    trip.TripID,
				ServiceID: trip.ServiceID,
				RouteID:   trip.RouteID,
			}

			/* stopTimeIxs, ok := stopTimesIxByTrip[trip.TripID]
			if ok {
				tripResponse.Schedules = make([]ScheduleResponse, 0, len(stopTimeIxs))
				for stopTimeIx := range stopTimeIxs {
					stopTime := stopTimes[stopTimeIx]
					tripResponse.Schedules = append(tripResponse.Schedules, ScheduleResponse{
						StopID:    stopTime.StopID,
						Arrival:   stopTime.Arrival,
						Departure: stopTime.Departure,
					})
				}
			} else { */
				tripResponse.Schedules = []ScheduleResponse{}
			// }
			resp = append(resp, tripResponse)
		}
		return resp
	} else {
		return []TripResponse{}
	}
}

func main() {
	stopTimes, stopTimesIxByTrip := getStopTimes()
	trips, tripsIxByRoute := getTrips()

	http.HandleFunc("/schedules/", func(w http.ResponseWriter, r *http.Request) {
		route := strings.Split(r.URL.Path, "/")[2]
		resp := buildTripResponse(route, stopTimes, stopTimesIxByTrip, trips, tripsIxByRoute)
		w.Header().Set("Content-Type", "application/json")
		json_resp, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("json error", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
		} else {
			w.Write(json_resp)
		}
	})
	log.Fatal(http.ListenAndServe(":4000", nil))
}

func getStopTimes() ([]StopTime, map[string][]int) {
	f, err := os.Open("./stop_times.big.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	start := time.Now()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	if records[0][0] != "trip_id" || records[0][3] != "stop_id" || records[0][1] != "arrival_time" || records[0][2] != "departure_time" {
		fmt.Println("stop_times.txt not in expected format:")
		for i, cell := range records[0] {
			fmt.Println(i, cell)
		}
		panic(1)
	}

	stopTimes := make([]StopTime, 0, 1_000_000)
	stsByTrip := make(map[string][]int)
	for i, rec := range records[1:] {
		trip := rec[0]
		sts, ok := stsByTrip[trip]
		if ok {
			stsByTrip[trip] = append(sts, i)
		} else {
			stsByTrip[trip] = []int{i}
		}
		stopTimes = append(stopTimes, StopTime{TripID: trip, StopID: rec[3], Arrival: rec[1], Departure: rec[2]})
	}
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("parsed", len(stopTimes), "stop times in", elapsed)

	return stopTimes, stsByTrip
}

func getTrips() ([]Trip, map[string][]int) {
	f, err := os.Open("./trips.big.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	start := time.Now()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	if records[0][2] != "trip_id" || records[0][0] != "route_id" || records[0][1] != "service_id" {
		fmt.Println("trips.txt not in expected format:")
		for i, cell := range records[0] {
			fmt.Println(i, cell)
		}
		panic(1)
	}

	trips := make([]Trip, 0, 70_000)
	tripsByRoute := make(map[string][]int)
	for i, rec := range records[1:] {
		route := rec[0]
		ts, ok := tripsByRoute[route]
		if ok {
			tripsByRoute[route] = append(ts, i)
		} else {
			tripsByRoute[route] = []int{i}
		}
		trips = append(trips, Trip{TripID: rec[2], RouteID: route, ServiceID: rec[1]})
	}
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("parsed", len(trips), "trips in", elapsed)

	return trips, tripsByRoute
}