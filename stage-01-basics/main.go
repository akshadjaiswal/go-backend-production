// package main is the entry point of every Go program.
// Every Go file starts with a package declaration.
// "main" is special — it tells Go: this is an executable program, not a library.
package main

// import brings in code from other packages.
// These are all from Go's standard library — no external dependencies needed.
import (
	"encoding/json" // for encoding Go structs into JSON
	"fmt"           // for printing to the console (like console.log in JS)
	"net/http"      // Go's built-in HTTP server — powerful enough for production
)

// Response is a struct — Go's way of defining a custom data type.
// Think of it like an object shape in TypeScript: { message: string, status: number }
// The backtick tags (json:"message") tell the JSON encoder what to name the fields in the output.
// Without tags, Go would use the field name as-is (e.g., "Message" with capital M).
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// healthHandler is a handler function — it handles HTTP requests to a specific route.
// Every handler in Go must have this exact signature:
//   - w http.ResponseWriter — you write your response into this (headers + body)
//   - r *http.Request      — this contains everything about the incoming request (URL, headers, body, method, etc.)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Set the Content-Type header so the client knows we're sending JSON.
	// This must be set BEFORE calling w.WriteHeader() or writing the body.
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code. 200 = OK.
	// http.StatusOK is a constant = 200. Using constants > magic numbers.
	w.WriteHeader(http.StatusOK)

	// json.NewEncoder(w) creates a JSON encoder that writes directly into the response.
	// .Encode(Response{...}) converts our struct to JSON and writes it to the response body.
	json.NewEncoder(w).Encode(Response{
		Message: "Server is running",
		Status:  http.StatusOK,
	})
}

// helloHandler demonstrates reading query parameters from the URL.
// Example: GET /hello?name=Akshad
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// r.URL.Query() parses the query string (?name=Akshad&age=25)
	// .Get("name") returns the value of the "name" param, or "" if not present
	name := r.URL.Query().Get("name")

	// If no name was passed, default to "World"
	if name == "" {
		name = "World"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// fmt.Sprintf works like template literals in JS: fmt.Sprintf("Hello, %s!", name) → "Hello, Akshad!"
	// %s = string placeholder (like %d for int, %f for float, %v for anything)
	json.NewEncoder(w).Encode(Response{
		Message: fmt.Sprintf("Hello, %s!", name),
		Status:  http.StatusOK,
	})
}

// main() is the entry point — Go starts executing your program from here.
func main() {
	// http.NewServeMux() creates a new router (multiplexer).
	// A mux maps URL paths → handler functions.
	// We use NewServeMux() instead of the default mux (http.DefaultServeMux) because:
	//   1. It's isolated — no risk of route conflicts with other packages
	//   2. Better for testing — you can create fresh muxes in tests
	mux := http.NewServeMux()

	// Register our handlers.
	// "GET /health" means: only match GET requests to /health
	// This method+path pattern is available in Go 1.22+ — older Go used just "/health"
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /hello", helloHandler)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	// http.ListenAndServe starts the HTTP server.
	// ":8080" means listen on all network interfaces, port 8080
	// mux is our router — it decides which handler to call for each request
	// This call BLOCKS — it runs forever until the program is killed.
	// If it returns, something went wrong, so we panic.
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
