package handlers

import (
	"Zoo_List/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NewPetRequest struct {
	Species    string      `json:"species"`
	Name       string      `json:"name"`
	SkillName  string      `json:"skill_name"`
	SkillLevel int         `json:"skill_level"`
	Stats      NewPetStats `json:"stats"`
	IsIdeal    bool        `json:"is_ideal"`
	IsMutant   bool        `json:"is_mutant"`
}

type NewPetStats struct {
	Loyalty  int `json:"loyalty"`
	Agility  int `json:"agility"`
	Stamina  int `json:"stamina"`
	Charm    int `json:"charm"`
	Instinct int `json:"instinct"`
}

func (h *Handler) CreatePet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	fmt.Printf("Creating pet for user ID: %d\n", userID)
	if r.Method == "POST" {
		var req NewPetRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			utils.WriteInvalidJSON(w)
			return
		}

		fmt.Printf("Received pet data: %+v\n", req)

		result, err := h.DB.Exec("INSERT INTO pets (owner_id, species, name, skill_name, skill_level, is_ideal, is_mutant) VALUES (?, ?, ?, ?, ?, ?, ?)", userID, req.Species, req.Name, req.SkillName, req.SkillLevel, req.IsIdeal, req.IsMutant)
		if err != nil {
			utils.WriteError500(w, "Error saving pet")
			return
		}

		petID, err := result.LastInsertId()
		if err != nil {
			utils.WriteError500(w, "Error getting pet ID")
			return
		}
		_, err = h.DB.Exec("INSERT INTO pet_stats (pet_id, loyalty, agility, stamina, instinct, charm) VALUES (?, ?, ?, ?, ?, ?)", petID, req.Stats.Loyalty, req.Stats.Agility, req.Stats.Stamina, req.Stats.Instinct, req.Stats.Charm)
		if err != nil {
			utils.WriteError500(w, "Error saving pet stats")
			return
		}
		response := map[string]interface{}{
			"message": "Pet created successfully",
			"pet_id":  petID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

type PetResponse struct {
	ID          int         `json:"id"`
	Species     string      `json:"species"`
	Name        string      `json:"name"`
	SkillName   string      `json:"skill_name"`
	SkillLevel  int         `json:"skill_level"`
	IsIdeal     bool        `json:"is_ideal"`
	IsMutant    bool        `json:"is_mutant"`
	Stats       NewPetStats `json:"stats"`
	AvailableIn int         `json:"available_in_hours,omitempty"`
	QueueSize   *int        `json:"queue_size,omitempty"`         
}

func (h *Handler) GetPets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	rows, err := h.DB.Query("SELECT p.id, p.species, p.name, p.skill_name, p.skill_level, p.is_ideal, p.is_mutant, ps.loyalty, ps.agility, ps.stamina, ps.charm, ps.instinct  FROM pets p JOIN pet_stats ps on p.id = ps.pet_id WHERE p.id NOT IN (SELECT pet_id FROM bookings WHERE user_id = ? AND status = 'active')", userID)
	if err != nil {
		utils.WriteDBError(w)
		return
	}
	defer rows.Close()

	var petList []PetResponse
	for rows.Next() {
		var pet PetResponse
		err = rows.Scan(&pet.ID, &pet.Species, &pet.Name, &pet.SkillName, &pet.SkillLevel, &pet.IsIdeal, &pet.IsMutant, &pet.Stats.Loyalty, &pet.Stats.Agility, &pet.Stats.Stamina, &pet.Stats.Charm, &pet.Stats.Instinct)
		if err != nil {
			utils.WriteDBError(w)
			return
		}
		pet.AvailableIn = h.calculatePetAvailability(pet.ID)

		petList = append(petList, pet)
	}
	if err = rows.Err(); err != nil {
		utils.WriteError500(w, "Error processing data")
		_ = fmt.Errorf("error processing data: %w", err)
		return
	}

	response := map[string]interface{}{
		"petList": petList,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetMyPets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	rows, err := h.DB.Query("SELECT p.id, p.species, p.name, p.skill_name, p.skill_level, p.is_ideal, p.is_mutant, ps.loyalty, ps.agility, ps.stamina, ps.charm, ps.instinct, (SELECT COUNT(*) FROM bookings WHERE pet_id = p.id AND status = 'active') as queue_size  FROM pets p JOIN pet_stats ps on p.id = ps.pet_id WHERE p.owner_id = ?", userID)
	if err != nil {
		utils.WriteDBError(w)
		return
	}
	defer rows.Close()

	var myPetList []PetResponse
	for rows.Next() {
		var pet PetResponse
		err = rows.Scan(&pet.ID, &pet.Species, &pet.Name, &pet.SkillName, &pet.SkillLevel, &pet.IsIdeal, &pet.IsMutant, &pet.Stats.Loyalty, &pet.Stats.Agility, &pet.Stats.Stamina, &pet.Stats.Charm, &pet.Stats.Instinct, &pet.QueueSize)
		if err != nil {
			utils.WriteDBError(w)
			return
		}
		pet.AvailableIn = h.calculatePetAvailability(pet.ID)

		myPetList = append(myPetList, pet)
	}
	if err = rows.Err(); err != nil {
		utils.WriteError500(w, "Error processing data")
		_ = fmt.Errorf("error processing data: %w", err)
		return
	}

	response := map[string]interface{}{
		"myPetList": myPetList,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

type QueueValue struct {
	Position     int       `json:"position"`
	UserNickname string    `json:"user_nickname"`
	CreatedAt    time.Time `json:"created_at"`
}

type FullPetResponse struct {
	PetResponse  PetResponse  `json:"petResponse"`
	CurrentQueue []QueueValue `json:"current_queue"`
}

func (h *Handler) GetPetByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	IDstr := strings.TrimPrefix(path, "/api/pets/")
	petID, err := strconv.Atoi(IDstr)
	if err != nil {
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	var fullPetInfo FullPetResponse
	pet := &fullPetInfo.PetResponse
	err = h.DB.QueryRow("SELECT p.id, p.species, p.name, p.skill_name, p.skill_level, p.is_ideal, p.is_mutant, ps.loyalty, ps.agility, ps.stamina, ps.charm, ps.instinct, (SELECT COUNT(*) FROM bookings WHERE pet_id = p.id AND status = 'active') as queue_size FROM pets p JOIN pet_stats ps on p.id = ps.pet_id WHERE p.id = ?", petID).Scan(&pet.ID, &pet.Species, &pet.Name, &pet.SkillName, &pet.SkillLevel, &pet.IsIdeal, &pet.IsMutant, &pet.Stats.Loyalty, &pet.Stats.Agility, &pet.Stats.Stamina, &pet.Stats.Charm, &pet.Stats.Instinct, &pet.QueueSize)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteNotFound(w, "Pet")
		} else {
			utils.WriteDBError(w)
		}
		return
	}

	if petID <= 0 {
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	rows, err := h.DB.Query("SELECT b.position, u.nickname, b.created_at FROM bookings b JOIN users u on b.user_id = u.id WHERE b.pet_id = ? AND b.status = 'active'", petID)
	if err != nil {
		utils.WriteDBError(w)
		return
	}
	defer rows.Close()

	var queue []QueueValue
	for rows.Next() {
		var queuePart QueueValue
		var createdAt string
		err = rows.Scan(&queuePart.Position, &queuePart.UserNickname, &createdAt)
		if err != nil {
			utils.WriteError500(w, "Error reading data")
			return
		}
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			queuePart.CreatedAt = parsedTime
		}
		queue = append(queue, queuePart)
	}

	if err = rows.Err(); err != nil {
		utils.WriteError500(w, "Error processing data")
		_ = fmt.Errorf("error processing data: %w", err)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    fullPetInfo,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) calculatePetAvailability(petID int) int {
	var queueSize int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE pet_id = ? AND status = 'active'", petID).Scan(&queueSize)
	if err != nil {
		return 0
	}

	var lastBooking *time.Time
	err = h.DB.QueryRow("SELECT last_successful_booking_date FROM pets WHERE id = ?", petID).Scan(&lastBooking)
	if err != nil {
		return queueSize * 48
	}

	if lastBooking == nil && queueSize == 0 {
		return 0
	}
	return calculateWaitTime(queueSize+1, lastBooking)
}

func (h *Handler) DeletePet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		utils.WriteError400(w, "Invalid URL format")
		return
	}

	petID, err := strconv.Atoi(parts[3])
	if err != nil || petID <= 0 {
		utils.WriteError400(w, "Invalid pet ID")
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		utils.WriteError500(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM pet_stats WHERE pet_id = ?", petID)
	if err != nil {
		utils.WriteError500(w, "Error deleting pet stats")
	}

	_, err = tx.Exec("DELETE FROM pets WHERE id = ?", petID)
	if err != nil {
		utils.WriteError500(w, "Error deleting pet")
	}

	err = tx.Commit()
	if err != nil {
		utils.WriteError500(w, "Failed to complete deletion")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Pet deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

type EditPetRequest struct {
	Species    *string       `json:"species,omitempty"`
	Name       *string       `json:"name,omitempty"`
	SkillName  *string       `json:"skill_name,omitempty"`
	SkillLevel *int          `json:"skill_level,omitempty"`
	Stats      *EditPetStats `json:"stats,omitempty"`
	IsIdeal    *bool         `json:"is_ideal,omitempty"`
	IsMutant   *bool         `json:"is_mutant,omitempty"`
}

type EditPetStats struct {
	Loyalty  *int `json:"loyalty,omitempty"`
	Agility  *int `json:"agility,omitempty"`
	Stamina  *int `json:"stamina,omitempty"`
	Charm    *int `json:"charm,omitempty"`
	Instinct *int `json:"instinct,omitempty"`
}

func (h *Handler) EditPet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	IDstr := strings.TrimSuffix(path, "/edit")
	IDstr = strings.TrimPrefix(IDstr, "/api")
	IDstr = strings.TrimPrefix(IDstr, "/pets/")
	petID, err := strconv.Atoi(IDstr)
	if err != nil || petID <= 0 {
		utils.WriteError400(w, "Invalid ID")
		return
	}

	var req EditPetRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteInvalidJSON(w)
		return
	}

	trx, err := h.DB.Begin()
	if err != nil {
		utils.WriteError500(w, "Failed to start transaction")
		return
	}
	defer trx.Rollback()

	var update []string
	var args []interface{}

	if req.Species != nil {
		update = append(update, "p.species=?")
		args = append(args, *req.Species)
	}
	if req.Name != nil {
		update = append(update, "p.name=?")
		args = append(args, *req.Name)
	}
	if req.SkillName != nil {
		update = append(update, "p.skill_name=?")
		args = append(args, *req.SkillName)
	}
	if req.SkillLevel != nil {
		update = append(update, "p.skill_level=?")
		args = append(args, *req.SkillLevel)
	}
	if req.Stats != nil {
		if req.Stats.Loyalty != nil {
			update = append(update, "ps.loyalty=?")
			args = append(args, *req.Stats.Loyalty)
		}
		if req.Stats.Agility != nil {
			update = append(update, "ps.agility=?")
			args = append(args, *req.Stats.Agility)
		}
		if req.Stats.Stamina != nil {
			update = append(update, "ps.stamina=?")
			args = append(args, *req.Stats.Stamina)
		}
		if req.Stats.Charm != nil {
			update = append(update, "ps.charm=?")
			args = append(args, *req.Stats.Charm)
		}
		if req.Stats.Instinct != nil {
			update = append(update, "ps.instinct=?")
			args = append(args, *req.Stats.Instinct)
		}
	}
	if req.IsIdeal != nil {
		update = append(update, "p.is_ideal=?")
		args = append(args, *req.IsIdeal)
	}
	if req.IsMutant != nil {
		update = append(update, "p.is_mutant=?")
		args = append(args, *req.IsMutant)
	}

	if len(update) == 0 {
		utils.WriteError400(w, "No fields to update")
		return
	}

	query := "UPDATE pets p JOIN pet_stats ps ON p.id = ps.pet_id SET " + strings.Join(update, ",") + " WHERE p.id = ?"
	args = append(args, petID)

	_, err = trx.Exec(query, args...)
	if err != nil {
		utils.WriteError500(w, "Error updating pet")
		return
	}

	err = trx.Commit()
	if err != nil {
		utils.WriteError500(w, "Failed to commit transaction")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Updated pet successfully",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}
