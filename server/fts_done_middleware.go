package server

import "net/http"

// checkIfRouteIsAllowedBeforeOrAfterFTS determines if a route is permitted based on its timing relative to FTS completion.
func checkIfRouteIsAllowedBeforeOrAfterFTS(isAllowedAfterFts, isAllowedBeforeFts bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
