package health

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Status struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func StartServer(db *sql.DB) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := Status{
			Status:   "ok",
			Database: "ok",
		}

		if err := db.Ping(); err != nil {
			status.Status = "degraded"
			status.Database = "unavailable"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	http.ListenAndServe(":8080", nil)
}
