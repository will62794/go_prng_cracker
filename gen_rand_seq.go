package main

/*
 * Go random number generation.
 */

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	// seq_length, _ := strconv.ParseInt(args[0], 10, 64)
	seed, _ := strconv.ParseInt(args[0], 10, 64)

	rand_source := rand.New(rand.NewSource(0))
	rand.Seed(seed)

	// var seq int64 = 0
	skips := 120
	jumps := 3

	for i := 0; i < skips; i++ {
		rand.Int()
	}

	for k := 0; k < jumps; k++ {
		// Generate sequence from seed.
		for i := int64(0); i < 18; i++ {
			val := rand.Intn(2)
			fmt.Printf("%d", val)
		}

		fmt.Printf(" ")

		// Skip some.
		skip_amt := 12 + rand_source.Intn(5)
		for i := 0; i < skip_amt; i++ {
			rand.Int()
		}
	}

}
