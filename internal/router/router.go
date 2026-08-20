package router

import (
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/handler"
)

func RegisterRoutes(
	mux *http.ServeMux,
	factHandler *handler.FactHandler,
	ruleHandler *handler.RuleHandler,
	resolutionHandler *handler.ResolutionHandler,
) {
	mux.HandleFunc(
		"GET /facts/{factInformationID}",
		factHandler.Get,
	)

	mux.HandleFunc(
		"POST /facts",
		factHandler.Create,
	)

	mux.HandleFunc(
		"GET /facts/details",
		factHandler.FindByEffectiveTime,
	)


	mux.HandleFunc(
		"POST /rules",
		ruleHandler.Create,
	)

	mux.HandleFunc(
		"GET /rules/find",
		ruleHandler.Find,
	)


	mux.HandleFunc(
		"POST /resolutions",
		resolutionHandler.Create,
	)

	mux.HandleFunc(
		"GET /resolutions/{resolutionID}",
		resolutionHandler.Get,
	)
}