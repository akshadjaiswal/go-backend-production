package main

import (
	"fmt"
	"net/http"

	"github.com/akshadjaiswal/go-backend-production/stage-02-routing/routes"
)

func main() {
	// routes.Setup() returns our fully configured Chi router with all routes registered
	r := routes.Setup()

	fmt.Println("Stage 02 — Server starting on http://localhost:8080")
	fmt.Println("Routes:")
	fmt.Println("  GET    /health")
	fmt.Println("  GET    /api/v1/users")
	fmt.Println("  POST   /api/v1/users")
	fmt.Println("  GET    /api/v1/users/{id}")
	fmt.Println("  PUT    /api/v1/users/{id}")
	fmt.Println("  DELETE /api/v1/users/{id}")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
