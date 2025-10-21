package middleware

import (
	"Zoo_List/internal/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func PetOwnerMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}
			fmt.Println("=== OwnerChecking STARTED ===")
			path := r.URL.Path
			var IDstr string
			if strings.Contains(path, "/delete") {
				IDstr = strings.TrimSuffix(path, "/delete")
			} else if strings.Contains(path, "/edit") {
				IDstr = strings.TrimSuffix(path, "/edit")
			} else if strings.Contains(path, "/up") {
				IDstr = strings.TrimSuffix(path, "/queue/up")
			} else if strings.Contains(path, "/down") {
				IDstr = strings.TrimSuffix(path, "/queue/down")
			} else if strings.Contains(path, "/complete") {
				IDstr = strings.TrimSuffix(path, "/complete")
			}

			IDstr = strings.TrimPrefix(IDstr, "/api")
			IDstr = strings.TrimPrefix(IDstr, "/pets/")

			petID, err := strconv.Atoi(IDstr)
			if err != nil {
				utils.WriteError400(w, "BAD_REQUEST")
				return
			}

			if petID <= 0 {
				utils.WriteError400(w, "Invalid pet ID")
				return
			}

			userID := r.Context().Value("user_id").(int)

			var ownerID int
			err = db.QueryRow("SELECT owner_id FROM pets WHERE id = ?", petID).Scan(&ownerID)
			if err != nil {
				if err == sql.ErrNoRows {
					utils.WriteError404(w, "Pet not found")
				} else {
					utils.WriteDBError(w)
				}
				return
			}
			
			if ownerID != userID {
				utils.WriteError403(w, "Access denied: not owner")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
