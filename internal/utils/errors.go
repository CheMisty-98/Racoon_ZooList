package utils

import (
	"log"
	"net/http"
	"runtime"
)

// Универсальная функция для логирования и возврата ошибок
func WriteError(w http.ResponseWriter, message string, status int) {
	_, file, line, _ := runtime.Caller(1)
	log.Printf("ERROR [%d] %s:%d - %s", status, file, line, message)
	http.Error(w, message, status)
}

// Конкретные ошибки 4xx
func WriteError400(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusBadRequest)
}

func WriteError401(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusUnauthorized)
}

func WriteError403(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusForbidden)
}

func WriteError404(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusNotFound)
}

func WriteError409(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusConflict)
}

func WriteError429(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusTooManyRequests)
}

// Серверные ошибки 5xx
func WriteError500(w http.ResponseWriter, message string) {
	WriteError(w, message, http.StatusInternalServerError)
}

// Специфичные ошибки с предопределенными сообщениями
func WriteInvalidJSON(w http.ResponseWriter) {
	WriteError400(w, "Invalid JSON")
}

func WriteDBError(w http.ResponseWriter) {
	WriteError500(w, "Database error")
}

func WriteNotFound(w http.ResponseWriter, resource string) {
	WriteError404(w, resource+" not found")
}

func WriteAlreadyExists(w http.ResponseWriter, resource string) {
	WriteError409(w, resource+" already exists")
}

func WriteUserNotFound(w http.ResponseWriter) {
	WriteError401(w, "User not found")
}

func WriteInvalidPassword(w http.ResponseWriter) {
	WriteError401(w, "Invalid password")
}
