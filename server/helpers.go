package server

/*
Code Ownership:
Finn - All functions
*/

import (
	"CTng/gossip"
	"fmt"
	"net/http"
	"strings"
)

// Given an http request, extract the gossip object ID parameters and return the ID
// Called in handleGossipObjectRequest. err is not nil when the request is missing these fields
func gossipIDFromParams(r *http.Request) (gossip.Gossip_object_ID, error) {
	gossipID := gossip.Gossip_object_ID{
		Application: r.FormValue("application"),
		Type:        r.FormValue("type"),
		Signer:      r.FormValue("signer"),
		Period:      r.FormValue("period"),
	}
	if gossipID.Application == "" || gossipID.Type == "" || gossipID.Signer == "" || gossipID.Period == "" {
		return gossipID, fmt.Errorf("Missing Parameters")
	}
	return gossipID, nil
}

// Gets the senderURl from an HTTP request.
// This information can be found in two locations typically, and the function searches these two places.
// Note that the function can potentially return Ipv6 addresses
func getSenderURL(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// Checks whether a parsed url equals the given URL, typically returned from getSenderURL above.
// Used within handleOwnerGossip() by the gossiper.
// This was a difficult task, and the below is not an eleagant implementation: it's more spaghetti
// It also hasn't been tested when working non-locally.
func isOwner(ownerURL string, parsedURL string) bool {
	// Aspects of this function may be wrong due to IPv6.
	if strings.Contains(parsedURL, "[::1]") {
		return true
	}
	// The below lines should remove the connection's port from the two URLs.
	ownerURL = strings.Split(ownerURL, ":")[0]
	parsedURL = strings.Split(parsedURL, ":")[0]
	if ownerURL == "localhost" || ownerURL == "[::1]" {
		if parsedURL == "localhost" || parsedURL == "[::1]" {
			return true
		}
	}
	return ownerURL == parsedURL
}
