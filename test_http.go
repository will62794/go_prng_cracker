package main

import "fmt"
import "net/http"
import "math/rand"

func main() {
	var seed int64 = 42

	// Seed.
	rand.Seed(seed)
	fmt.Printf("Seeded 'rand' with %d\n", seed)

	for i := 0; i < 6; i++ {
		fmt.Printf("Generation %d: %d\n", i, rand.Int())
	}

	fmt.Println("-----")

	// Re-seed.
	rand.Seed(seed)
	fmt.Printf("Seeded 'rand' with %d\n", seed)

	fmt.Printf("Pre  HTTP request: %d\n", rand.Int())
	http.Get("http://google.com")
	fmt.Printf("Post HTTP request: %d\n", rand.Int())
}
