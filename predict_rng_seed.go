package main

/*
	Cracking Go "math/rand" random number generation.

	The seed for a "math/rand" random number source is a 64-bit integer, but is
	actually taken  modulo 2^31-1. (https://github.com/golang/go/blob/master/src/math/rand/rng.go#L205)
	Therefore, there are only that many possible unique seeds. If we observe an (initial)
	sequence of random numbers generated from a Go Rand source, we can try to
	check (brute force) all possible seeds to see if they produce the same
	sequence of numbers we observed. Observing a reasonable length sequence,
	say 50 numbers, should provide a pretty good likelihood that a seed that
	produces that exact sequence is the right one.

	Will Schultz, June 2017
*/

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	SEQ_LEN    = 31   // length of observed sequence
	GEN_LENGTH = 2000 // length of sequence to generate from seed.
	GEN_SKIPS  = 100  // number of RNG generations to skip initially
	RANDN      = 2    // argument to rand.Intn when checking for seed
)

var (
	// Observed sequence of random numbers to match
	// goal_seq = [SEQ_LEN]int64 {0,1,1,0,1,0,1,0,1,1,1,1,1,0,1,1,0,0,0,1,1,1,0,1,0,1,1,1,1,0,1,1,1,0,1} //,0,1,1,1,1,1,0,1,0,1,1,0,1,0,0} //mongobucks coin tosses
	// observed = [SEQ_LEN]int64 {1,1,0,0,0,0,0,1,1,0,0,1,1,1,0,1,0,1,0,1,0,0,1,0,0,0,1,1,0,0,1}
	// goal_seq = [SEQ_LEN]int64 {1,0,0,1,1,1,0,1,0,1,1,0,0,0,0,0,0,0,0,1,1,0,1,1,0,1,1,1,1,0,0,0,0,1,0,1,1,0,0,0,0,0,0,1,0,1,1,0,1,1}
	// observed = "1100000110011101010100100011001"

	observed = "0101111101100010010110111111010011000111000101110001110011110110001100111010010110010010111011011111"
	// observed = "01001000110011001001010000011010001110"
	// observed2 = "0010101111101111"
	// observed3 = ""
)

// Check if a given seed generates the goal sequence.
func check_seed(seed int64, rand_source *rand.Rand, seq [SEQ_LEN]int64) bool {
	rand_source.Seed(seed)
	var genSeq bytes.Buffer

	// Skip some of the initial generations.
	for i := 0; i < GEN_SKIPS; i++ {
		rand_source.Intn(RANDN)
	}
	ta1 := time.Now().UnixNano()

	for i := 0; i < GEN_LENGTH; i++ {
		val := rand_source.Intn(RANDN)
		// seq[i]=int64(val)
		if val == 1 {
			genSeq.WriteString("1")
		} else {
			genSeq.WriteString("0")
		}
	}
	ta2 := time.Now().UnixNano()
	fmt.Println(ta2 - ta1)

	genSeqStr := genSeq.String()
	// fmt.Println(genSeqStr)
	// fmt.Println(observed)

	// Check if this seed likely produced the goal sequence.
	t1 := time.Now().UnixNano()
	p1 := strings.Index(genSeqStr, observed[:60])
	t2 := time.Now().UnixNano()
	fmt.Println(t2 - t1)
	fmt.Println("---\n\n\n")
	if p1 > 0 {
		fmt.Printf("found index: %d\n", p1)
		return true
	}

	p2 := strings.Index(genSeqStr, observed[40:100])
	if p2 > 0 {
		return true
	}

	p3 := strings.Index(genSeqStr, observed[20:80])
	if p3 > 0 {
		return true
	}

	// p2 := strings.Index(genSeqStr[p1+len(observed):], observed2)
	// if p2 == -1 {
	// 	return false
	// }

	// p3 := strings.Index(genSeqStr[p2+len(observed2):], observed3)
	// if p3 == -1 {
	// 	return false
	// }

	return false
}

// Search for a seed in a specified range that generates the goal sequence. Sends the seed value
// on 'found_seed_ch' when it finds it.
func seed_search(part_id int64, lo int64, hi int64, found_seed_ch chan int64) {
	var printFreq int64 = 100000

	// Create randomness source.
	rand_source := rand.New(rand.NewSource(0))
	var seq [SEQ_LEN]int64

	var currSeed int64
	for currSeed = lo; currSeed < hi; currSeed++ {
		// Print progress
		if (currSeed % printFreq) == 0 {
			progress := float64(currSeed-lo) / float64(hi-lo) * 100
			fmt.Printf("(Partition %2d) Currently checking seed %d (%f %%)\n", part_id, currSeed, progress)
		}

		found := check_seed(currSeed, rand_source, seq)
		if found {
			found_seed_ch <- currSeed
		}
	}
}

func main() {

	args := os.Args[1:]
	arg0, _ := strconv.ParseInt(args[0], 10, 64)

	var seedMin int64 = 0
	var seedMax int64 = (1 << 31) - 1

	// Number of partitions to break the seed space up into.
	var num_partitions int64 = arg0
	var partition_size int64 = (seedMax - seedMin) / num_partitions

	// Limit CPU Core Utilization.
	runtime.GOMAXPROCS(int(num_partitions))

	// Channel that will receive the found seed value.
	found_seed_ch := make(chan int64)

	var start = time.Now().Unix()
	var p int64
	for p = 0; p < num_partitions; p++ {
		lo := (partition_size * p) + seedMin
		hi := (partition_size * (p + 1)) + seedMin
		fmt.Printf("-- Partition %d (%d, %d) \n", p, lo, hi)
		// if p == 1 {
		// 	lo = 505747691
		// 	hi = 505747693
		// }
		go seed_search(p, lo, hi, found_seed_ch)
	}

	for {
		seedFound := <-found_seed_ch
		var end = time.Now().Unix()
		fmt.Printf("\n====> Found a likely seed in %d seconds. Seed is %d\n\n", end-start, seedFound)
	}
}
