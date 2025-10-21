package main

import (
	"Zoo_List/internal/config"
	"Zoo_List/internal/database"
	"Zoo_List/internal/handlers"
	"Zoo_List/internal/middleware"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()
	db, err := database.ConnectDb(*cfg)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("DB connected successfully!")
	}

	handler := &handlers.Handler{DB: db}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/api/register", handler.Register)
	http.HandleFunc("/api/login", handler.Login)
	http.Handle("/api/pets", middleware.AuthMiddleware(http.HandlerFunc(handler.CreatePet)))
	http.Handle("/api/bookings/my", middleware.AuthMiddleware(http.HandlerFunc(handler.GetMyBooking)))
	http.Handle("/api/pets/list", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPets)))
	http.Handle("/api/pets/my", middleware.AuthMiddleware(http.HandlerFunc(handler.GetMyPets)))
	http.Handle("/api/pets/{id}", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPetByID)))
	http.Handle("/api/pets/{id}/book", middleware.AuthMiddleware(http.HandlerFunc(handler.CreateBooking)))
	http.Handle("/api/pets/{id}/cancel", middleware.AuthMiddleware(http.HandlerFunc(handler.CancelBooking)))
	http.Handle("/api/pets/{id}/delete", middleware.AuthMiddleware(middleware.PetOwnerMiddleware(db)(http.HandlerFunc(handler.DeletePet))))
	http.Handle("/api/pets/{id}/edit", middleware.AuthMiddleware(middleware.PetOwnerMiddleware(db)(http.HandlerFunc(handler.EditPet))))
	http.Handle("/api/pets/{id}/queue", middleware.AuthMiddleware(http.HandlerFunc(handler.QueueView)))
	http.Handle("/api/pets/{id}/queue/{direction}", middleware.AuthMiddleware(middleware.PetOwnerMiddleware(db)(http.HandlerFunc(handler.MoveInQueue))))
	http.Handle("/api/pets/{id}/complete", middleware.AuthMiddleware(middleware.PetOwnerMiddleware(db)(http.HandlerFunc(handler.CompleteBooking))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	fmt.Printf("Server starting on port %s...\n", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, nil))
}
