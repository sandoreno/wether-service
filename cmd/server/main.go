package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/sandoreno/wether-service/cmd/internal/client/http/geocoding"
	"github.com/sandoreno/wether-service/cmd/internal/client/http/open_meteo"
	"log"
	"net/http"
	"sync"
	"time"
)

const httpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		fmt.Printf("Requested city: %s\n", city)

		geoRes, err := geocodingClient.GetCoords(city)
		if err != nil {
			log.Println(err)
		}

		openMetRes, err := openMeteoClient.GetTemperature(geoRes.Latitude, geoRes.Longitude)
		if err != nil {
			log.Println(err)
			return
		}

		row, err := json.Marshal(openMetRes)
		if err != nil {
			log.Println(err)
		}

		_, err = w.Write(row)
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		fmt.Println("starting http server on " + httpPort)
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()

		fmt.Printf("starting job: %v\n", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()

}

func initJobs(sheduler gocron.Scheduler) ([]gocron.Job, error) {
	// add a job to the scheduler
	j, err := sheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("Hello")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}
