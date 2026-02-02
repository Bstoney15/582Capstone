package routes

import (
	"net/http"
)

/*
All api routes must be defined in this file. All routes must also be prefixed with api (/api/route)
try to group routes by functionality and authentication requirements

*/
func AddRoutesToServer(s *http.ServeMux) {

	// No Auth Routes
	s.HandleFunc("/api/health", HealthCheckHandler)


	// Developer and above Routes

	
	// Admin Only Routes
}