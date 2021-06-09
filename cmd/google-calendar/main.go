package main

import (
	"github.com/tktip/google-calendar/internal/api"
	"github.com/tktip/google-calendar/pkg/healthcheck"
)

func main() {
	//Starting health check
	go healthcheck.StartHealthService()

	api.ListenAndServe()
}
