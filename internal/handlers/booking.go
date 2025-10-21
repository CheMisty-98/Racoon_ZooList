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

type BookingRequest struct {
	PetID int `json:"pet_id"`
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== CreateBooking STARTED ===")
	var req BookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteInvalidJSON(w)
		return
	}

	if req.PetID <= 0 {
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	var isExist bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM pets WHERE id = ?)", req.PetID).Scan(&isExist)
	if err != nil {
		utils.WriteDBError(w)
		return
	}

	if !isExist {
		utils.WriteNotFound(w, "Pet")
		return
	}

	userID := r.Context().Value("user_id").(int)
	var isBooking bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM bookings WHERE pet_id = ? AND user_id = ? AND status = 'active')", req.PetID, userID).Scan(&isBooking)
	if err != nil {
		utils.WriteDBError(w)
		return
	}
	if isBooking {
		utils.WriteAlreadyExists(w, "Booking")
		return
	}

	var newPosition int

	err = h.withPetLock(req.PetID, func() error {
		fmt.Println("Блокировка работает!")
		var queueLen int
		err = h.DB.QueryRow("SELECT COALESCE(MAX(position),0) FROM bookings WHERE pet_id = ? AND status = 'active'", req.PetID).Scan(&queueLen)
		if err != nil {
			return fmt.Errorf("queue error: %w", err)
		}
		newPosition = queueLen + 1
		_, err = h.DB.Exec("INSERT INTO bookings(pet_id, user_id, position, status) Values(?, ?, ?, ?) ", req.PetID, userID, newPosition, "active")
		if err != nil {
			return fmt.Errorf("insert error: %w", err)
		}
		return nil
	})

	if err != nil {
		if err.Error() == "lock timeout" {
			utils.WriteError429(w, "System busy, try again later")
		} else {
			utils.WriteError500(w, "Server error")
		}
		return
	}

	fmt.Printf("Booking created successfully. PetID: %d, UserID: %d\n", req.PetID, userID)

	response := map[string]interface{}{
		"message":  "Booking created successfully",
		"position": newPosition,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)

	fmt.Println("=== CreateBooking COMPLETED ===")

}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== CancelBooking STARTED ===")

	path := r.URL.Path
	petIDStr := strings.TrimPrefix(path, "/api/pets/")
	petIDStr = strings.TrimSuffix(petIDStr, "/cancel")
	petID, err := strconv.Atoi(petIDStr)
	if err != nil || petID <= 0 {
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	userID := r.Context().Value("user_id").(int)

	var bookingID int
	var bookingPosition int
	err = h.DB.QueryRow(
		"SELECT id, position FROM bookings WHERE pet_id = ? AND user_id = ? AND status = 'active'",
		petID, userID,
	).Scan(&bookingID, &bookingPosition)

	if err != nil {
		utils.WriteNotFound(w, "Active booking for this pet")
		return
	}

	err = h.withPetLock(petID, func() error {
		fmt.Println("Блокировка работает!")
		_, err = h.DB.Exec("DELETE FROM bookings WHERE id = ?", bookingID)
		if err != nil {
			return fmt.Errorf("delete bookings error: %w", err)
		}
		_, err = h.DB.Exec("UPDATE bookings SET position = position - 1 WHERE pet_id = ? AND position > ? AND status = 'active'", petID, bookingPosition)
		if err != nil {
			return fmt.Errorf("update position error: %w", err)
		}
		return nil
	})

	if err != nil {
		if err.Error() == "lock timeout" {
			utils.WriteError429(w, "System busy, try again later")
		} else {
			utils.WriteError500(w, "Server error")
		}
		return
	}

	fmt.Printf("Booking canceled successfully. PetID: %d, UserID: %d\n", petID, userID)

	response := map[string]interface{}{
		"message": "Booking cancelled successfully",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
	fmt.Println("=== CancelBooking COMPLETED ===")
}

type MyBookingResopnce struct {
	BookingID         int        `json:"booking_id"`
	PetID             int        `json:"pet_id"` 
	BookingPosition   int        `json:"booking_position"`
	BookingCreatedAt  time.Time  `json:"booking_created_at"`
	PetName           string     `json:"pet_name"`
	PetSpecies        string     `json:"pet_species"`
	PetSkillName      string     `json:"pet_skill_name"`
	PetSkillLevel     int        `json:"pet_skill_level"`
	EstimatedWaitTime int        `json:"estimated_wait_time"`
	LastBookingDate   *time.Time `json:"-"`
}

func (h *Handler) GetMyBooking(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== GettingBooking STARTED ===")
	userID := r.Context().Value("user_id").(int)
	fmt.Printf("GetMyBooking STARTED for user %d\n", userID)

	rows, err := h.DB.Query("SELECT b.id, b.position, b.created_at, p.id as pet_id, p.name, p.species, p.skill_name, p.skill_level, p.last_successful_booking_date FROM bookings b JOIN pets p ON b.pet_id = p.id WHERE b.user_id = ? AND b.status = 'active' ORDER BY b.created_at DESC", userID)
	if err != nil {
		utils.WriteDBError(w)
		_ = fmt.Errorf("DB query error: %w", err)
		return
	}
	defer rows.Close()
	fmt.Println("Query executed successfully")

	var myBooking []MyBookingResopnce
	for rows.Next() {
		var booking MyBookingResopnce
		fmt.Println("Scanning now...")
		var createdAt string
		var lastBookingDate *string
		err = rows.Scan(
			&booking.BookingID,       
			&booking.BookingPosition, 
			&createdAt,               
			&booking.PetID,           
			&booking.PetName,         
			&booking.PetSpecies,      
			&booking.PetSkillName,    
			&booking.PetSkillLevel,   
			&lastBookingDate,         
		)
		if err != nil {
			utils.WriteError500(w, "Error reading data")
			_ = fmt.Errorf("error reading data: %w", err)
			return
		}

		fmt.Printf("Successfully scanned booking ID: %d, Pet ID: %d\n", booking.BookingID, booking.PetID)

		var parsedTime time.Time
		if parsedTime, err = time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			booking.BookingCreatedAt = parsedTime
		}

		if lastBookingDate != nil {
			parsedTime, err = time.Parse("2006-01-02 15:04:05", *lastBookingDate)
			if err == nil {
				booking.LastBookingDate = &parsedTime
			}
		}

		booking.EstimatedWaitTime = calculateWaitTime(booking.BookingPosition, booking.LastBookingDate)
		myBooking = append(myBooking, booking)
	}

	if err = rows.Err(); err != nil {
		utils.WriteError500(w, "Error processing data")
		_ = fmt.Errorf("error processing data: %w", err)
		return
	}

	fmt.Printf("GetMyBooking: found %d bookings for user %d\n", len(myBooking), userID)

	response := map[string]interface{}{
		"my_booking": myBooking,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
	fmt.Println("=== GettingBooking COMPLETED ===")
}

func calculateWaitTime(position int, booking *time.Time) int {
	if booking == nil {
		return (position - 1) * 48
	}

	hoursSinceLastBooking := time.Since(*booking).Hours()

	remainingCooldown := 48 - hoursSinceLastBooking
	if remainingCooldown < 0 {
		remainingCooldown = 0
	}

	totalWaitTime := (position-1)*48 + int(remainingCooldown)
	return totalWaitTime
}

func (h *Handler) withPetLock(petID int, operation func() error) error {
	lockName := "pet_queue_" + strconv.Itoa(petID)
	var isTimeOut int
	defer h.DB.QueryRow("SELECT RELEASE_LOCK(?)", lockName)
	err := h.DB.QueryRow("SELECT GET_LOCK(?, 10)", lockName).Scan(&isTimeOut)
	if err != nil {
		return fmt.Errorf("DB error: %w", err)
	}
	if isTimeOut == 1 {
		return operation()
	} else if isTimeOut == 0 {
		return fmt.Errorf("lock timeout")
	} else {
		return fmt.Errorf("lock error")
	}
}
