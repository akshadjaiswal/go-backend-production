package main

import (
	"fmt"
	"net/http"

	"github.com/akshadjaiswal/go-backend-production/stage-03-middleware/routes"
)

func main() {
	r := routes.Setup()

	fmt.Println("Stage 03 — Server starting on http://localhost:8080")
	fmt.Println("Middleware: RequestID → Logger → CORS → AuthGuard (on /api/v1/*)")
	fmt.Println("")
	fmt.Println("Routes:")
	fmt.Println("  GET    /health              (no auth)")
	fmt.Println("  GET    /api/v1/users        (requires X-API-Key header)")
	fmt.Println("  POST   /api/v1/users        (requires X-API-Key header)")
	fmt.Println("  GET    /api/v1/users/{id}   (requires X-API-Key header)")
	fmt.Println("  PUT    /api/v1/users/{id}   (requires X-API-Key header)")
	fmt.Println("  DELETE /api/v1/users/{id}   (requires X-API-Key header)")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
