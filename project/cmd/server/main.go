package main

import (
	"log"
	"net/http"
	"project/internal/web"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

	})

	srv, err := web.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Routes()))

	// res, err := http.NewRequest("GET", "https://api.myinvestor.es/myinvestor-server/rest/public/fondos/find-fondos?tipo=TODOS", nil)
	// fmt.Println(res)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
