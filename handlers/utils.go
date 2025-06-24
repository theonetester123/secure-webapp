package handlers

import (
	"net/http"
	"secure-webapp/models"
)

// Helper function for session management
func getOrCreateSession(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// Create new session
		sessionID := models.GenerateID()
		userID := models.GenerateID()

		models.SetSession(sessionID, userID)

		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		})

		return sessionID
	}

	// Validate existing session
	sessionID := cookie.Value
	_, exists := models.GetSession(sessionID)
	if !exists {
		// Session invalid, create new one
		userID := models.GenerateID()
		models.SetSession(sessionID, userID)
	}

	return sessionID
}
