package router

import (
	"net/http"

	"github.com/arbazshaikh150/TimeCourt/internal/handler"
)

func RegisterRoutes(
	mux *http.ServeMux,
	factHandler *handler.FactHandler,
	ruleHandler *handler.RuleHandler,
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

	// Rule handler
	mux.HandleFunc(
		"POST /rules",
		ruleHandler.Create,
	)

	mux.HandleFunc(
		"GET /rules/find",
		ruleHandler.Find,
	)
}
