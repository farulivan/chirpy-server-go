package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Reset is only allowed in dev environment.", nil)
		return
	}

	cfg.fileserverHits.Store(0)
	if err := cfg.db.DeleteAllUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete all users", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Hits reset to 0 and all users deleted"})
}
