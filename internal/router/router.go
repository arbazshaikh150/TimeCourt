package router

import (
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/handler"
)

func RegisterRoutes(
	mux *http.ServeMux,
	factHandler *handler.FactHandler,
) {
	mux.HandleFunc(
		"GET /facts/{factInformationID}",
		factHandler.Get,
	)

	mux.HandleFunc(
		"POST /facts",
		factHandler.Create,
	)
}
