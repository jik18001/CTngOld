package server

import (
	"CTng/gossip"
	"fmt"
	"net/http"
	"strings"
)

func gossipIDFromParams(r *http.Request) (gossip.Gossip_object_ID, error) {
	// Get the ID from the request.
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

func getSenderURL(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// Rules: If localhost, call it true.
// Otherwise compare the pre-port part of the url to see if they match.
func isOwner(ownerURL string, parsedURL string) bool {
	//aspects of this function may be wrong due to IPv6.
	if strings.Contains(parsedURL, "[::1]") {
		return true
	}
	ownerURL = strings.Split(ownerURL, ":")[0]
	parsedURL = strings.Split(parsedURL, ":")[0]
	if ownerURL == "localhost" || ownerURL == "[::1]" {
		if parsedURL == "localhost" || parsedURL == "[::1]" {
			return true
		}
	}
	return ownerURL == parsedURL
}
