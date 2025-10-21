package handlers

import (
	"Zoo_List/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type QueueData struct {
	BookingID int    `json:"booking_id"`
	UserID    int    `json:"user_id"`
	Position  int    `json:"position"`
	Nickname  string `json:"nickname"`
}

type QueueResponse struct {
	ID                    int        `json:"id"`
	UserID                int        `json:"user_id"`
	PetID                 int        `json:"pet_id"`
	Position              int        `json:"position"`
	CreatedAt             time.Time  `json:"created_at"`
	Nickname              string     `json:"nickname"`
	LastBookingDate       *time.Time `json:"last_booking_date"`
	HoursSinceLastBooking *int       `json:"hours_since_last_booking"`
	ReadyIn               int        `json:"ready_in"`
}

func (h *Handler) QueueView(w http.ResponseWriter, r *http.Request) {
	petID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Printf("Invalid pet ID: %v\n", err)
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	fmt.Printf("Processing queue for petID: %d\n", petID)

	rows, err := h.DB.Query("SELECT b.id, b.user_id, b.pet_id, b.position, b.created_at, u.nickname, p.last_successful_booking_date, TIMESTAMPDIFF(HOUR, p.last_successful_booking_date, NOW()) as hours_since_last_booking FROM bookings b JOIN users u on b.user_id = u.id JOIN pets p ON b.pet_id = p.id WHERE b.pet_id = ? AND b.status = 'active'ORDER BY position", petID)
	if err != nil {
		utils.WriteError500(w, "Error getting queue: "+err.Error())
		return
	}
	defer rows.Close()

	var bookings []QueueResponse
	for rows.Next() {
		var booking QueueResponse
		var createdAt string
		var lastBookingDate *string
		var hoursSinceLastBooking *int

		err = rows.Scan(&booking.ID, &booking.UserID, &booking.PetID, &booking.Position, &createdAt, &booking.Nickname, &lastBookingDate, &hoursSinceLastBooking)
		if err != nil {
			utils.WriteError500(w, "Error scanning booking: "+err.Error())
			return
		}

		if hoursSinceLastBooking != nil && *hoursSinceLastBooking < 48 {
			booking.ReadyIn = 48 - *hoursSinceLastBooking
		} else {
			booking.ReadyIn = 0
		}

		if booking.Position != 1 {
			booking.ReadyIn = booking.ReadyIn + (booking.Position-1)*48
		}

		if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			booking.CreatedAt = parsedTime
		}

		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		utils.WriteError500(w, "Error processing queue: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

func (h *Handler) MoveInQueue(w http.ResponseWriter, r *http.Request) {
	var user1MovementData, user2MovementData QueueData
	path := r.URL.Path
	parts := strings.Split(path, "/")

	petID, _ := strconv.Atoi(parts[3])
	direction := parts[5]

	if direction != "up" && direction != "down" {
		utils.WriteError400(w, "Invalid direction. Use 'up' or 'down'")
		return
	}

	var requestData struct {
		BookingID int `json:"bookingId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		utils.WriteInvalidJSON(w)
		return
	}

	err := h.DB.QueryRow("SELECT b.position, b.id, b.user_id, u.nickname FROM bookings b JOIN users u ON b.user_id = u.id WHERE b.id = ? AND b.status = 'active'", requestData.BookingID).Scan(&user1MovementData.Position, &user1MovementData.BookingID, &user1MovementData.UserID, &user1MovementData.Nickname)
	if err != nil {
		utils.WriteDBError(w)
		return
	}

	if direction == "up" && user1MovementData.Position == 1 {
		utils.WriteError400(w, "Already at the top")
		return
	}

	var maxPosition int
	err = h.DB.QueryRow("SELECT MAX(position) FROM bookings WHERE pet_id = ? AND status = 'active'", petID).Scan(&maxPosition)
	if direction == "down" && user1MovementData.Position == maxPosition {
		utils.WriteError400(w, "Already at the bottom")
		return
	}

	if direction == "up" {
		user2MovementData.Position = user1MovementData.Position - 1
	} else {
		user2MovementData.Position = user1MovementData.Position + 1
	}
	err = h.DB.QueryRow("SELECT b.id, b.user_id, u.nickname FROM bookings b JOIN users u ON b.user_id = u.id WHERE b.position = ? AND b.pet_id = ? AND b.status = 'active'", user2MovementData.Position, petID).Scan(&user2MovementData.BookingID, &user2MovementData.UserID, &user2MovementData.Nickname)
	if err != nil {
		utils.WriteDBError(w)
		return
	}

	trx, err := h.DB.Begin()
	if err != nil {
		utils.WriteError500(w, "Failed to start transaction")
		return
	}
	defer trx.Rollback()

	_, err = trx.Exec("UPDATE bookings SET position = -1 WHERE id = ?", user1MovementData.BookingID)

	_, err = trx.Exec("UPDATE bookings SET position = ? WHERE id = ?", user1MovementData.Position, user2MovementData.BookingID)

	_, err = trx.Exec("UPDATE bookings SET position = ? WHERE id = ?", user2MovementData.Position, user1MovementData.BookingID)
	if err != nil {
		fmt.Printf("SQL Error: %v\n", err)
		utils.WriteError500(w, "Failed to update queue positions")
		return
	}

	err = trx.Commit()
	if err != nil {
		utils.WriteError500(w, "Failed to commit transaction")
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Queue updated position successfully",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)

}

func (h *Handler) CompleteBooking(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CompleteBooking started")

	petID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		fmt.Printf("Invalid pet ID: %v\n", err)
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	fmt.Printf("Processing completion for petID: %d\n", petID)

	var hoursSinceLastBooking *int
	err = h.DB.QueryRow(`
        SELECT TIMESTAMPDIFF(HOUR, last_successful_booking_date, NOW()) 
        FROM pets WHERE id = ?
    `, petID).Scan(&hoursSinceLastBooking)

	if err != nil {
		fmt.Printf("Error checking pet readiness: %v\n", err)
		utils.WriteError500(w, "Error checking pet readiness")
		return
	}

	fmt.Printf("Hours since last booking: %v\n", hoursSinceLastBooking)

	if hoursSinceLastBooking != nil && *hoursSinceLastBooking < 48 {
		fmt.Printf("Pet not ready yet: %d hours since last booking\n", *hoursSinceLastBooking)
		utils.WriteError400(w, "Pet is not ready yet")
		return
	}

	var firstBookingExists bool
	err = h.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM bookings 
            WHERE pet_id = ? AND position = 1 AND status = 'active'
        )
    `, petID).Scan(&firstBookingExists)

	if err != nil {
		fmt.Printf("Error checking first booking: %v\n", err)
		utils.WriteError500(w, "Error checking booking")
		return
	}

	if !firstBookingExists {
		fmt.Println("No active booking to complete")
		utils.WriteError400(w, "No active booking to complete")
		return
	}

	fmt.Println("All checks passed, proceeding with completion...")

	result, err := h.DB.Exec(`
        UPDATE bookings 
        SET status = 'completed' 
        WHERE pet_id = ? AND position = 1 AND status = 'active'
    `, petID)

	if err != nil {
		utils.WriteError500(w, "Error completing booking: "+err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.WriteError400(w, "No active booking to complete")
		return
	}

	_, err = h.DB.Exec(`
        UPDATE pets 
        SET last_successful_booking_date = NOW() 
        WHERE id = ?
    `, petID)

	if err != nil {
		utils.WriteError500(w, "Error updating pet: "+err.Error())
		return
	}

	_, err = h.DB.Exec(`
        UPDATE bookings 
        SET position = position - 1 
        WHERE pet_id = ? AND status = 'active' AND position > 1
    `, petID)

	if err != nil {
		utils.WriteError500(w, "Error updating queue: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Booking completed successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
