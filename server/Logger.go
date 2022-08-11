package server

/*
Code Ownership:
Marcus - Function and object definitions.

___
For now, this file represents the functionality required of a ctng logger.
It should be implemented in the future. Currently, we utilize fakelogger (in the testData folder) for testing.

*/

import (
	"net/http"
)

// Struct for JSON STH object
type STH struct {
	date string `json:"date"`
	tree string `json:"tree"`
}

// Struct for JSON Entries object
type Entries struct {
	date         string `json:"date"`
	certificates string `json:"certificates"`
}

type logger interface {

	// Response to entering the 'base page' of a logger.
	// (May be completely unnecessary, or we may need to create a generic page for all entities)
	homePage()

	// Return the STH for the given date.
	// ctng/v1/get-sth/{date}
	returnSTH(w http.ResponseWriter, r *http.Request)

	// Return a list of encoded certificates registered on the given date.
	// ctng/v1/get-entries/{date}
	returnEntries(w http.ResponseWriter, r *http.Request)
}
