package models

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Nickname     string    `json:"nickname"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Pet struct {
	ID                        int       `json:"id"`
	OwnerID                   int       `json:"owner_id"`
	Species                   string    `json:"species"`
	Name                      string    `json:"name"`
	SkillName                 string    `json:"skill_name"`
	SkillLevel                int       `json:"skill_level"`
	IsIdeal                   bool      `json:"is_ideal"`
	IsMutant                  bool      `json:"is_mutant"`
	LastSuccessfulBookingDate time.Time `json:"last_successful_booking_date"`
	CreatedAt                 time.Time `json:"created_at"`
}

type PetStats struct {
	PetID    int `json:"pet_id"`
	Loyalty  int `json:"loyalty"`
	Agility  int `json:"agility"`
	Stamina  int `json:"stamina"`
	Instinct int `json:"instinct"`
	Charm    int `json:"charm"`
}

const (
	BookingStatusActive    = "active"
	BookingStatusCompleted = "completed"
	BookingStatusCancelled = "cancelled"
)

type Booking struct {
	ID        int       `json:"id"`
	PetID     int       `json:"pet_id"`
	UserID    int       `json:"user_id"`
	Position  float64   `json:"position"`
	Status    string    `json:"status"` // "active", "completed", "cancelled"
	CreatedAt time.Time `json:"created_at"`
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}
