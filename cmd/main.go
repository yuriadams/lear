package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yuriadams/lear/internal/controller"
	"github.com/yuriadams/lear/internal/repository"
	"github.com/yuriadams/lear/internal/usecase"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bookRepo := repository.NewBookRepository(db)
	bookUsecase := usecase.NewBookUsecase(bookRepo)
	bookHandler := controller.NewBookHandler(bookUsecase)

	router := mux.NewRouter()
	router.HandleFunc("/", bookHandler.Index).Methods("GET")
	router.HandleFunc("/books/{id:[0-9]+}", bookHandler.Show).Methods("GET")
	router.HandleFunc("/books/{id:[0-9]+}/analyze", bookHandler.StreamAnalysis).Methods("GET")

	fmt.Println("Server running on :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
