package main

/*
	Generate a lookup table that maps Go random number generator seed initial sequences to the seed
	that produced them.

	Will Schultz, June 2017
*/


import (
	"math/rand"
	"fmt"
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
	"runtime"
)

const(
	SEQ_LEN = 31 // length of observed sequence
	RANDN = 2 // argument to rand.Intn when checking for seed
	dbstring = "localhost:28000"
	collName = "mongobucks"
	dbName = "test"
)

type SeedRecord struct {
		Id int64 `json:"id" bson:"_id,omitempty"`
        Seed int64
}

// Given a seed, generate sequence of 1 bit values of length SEQ_LEN.
// Returns the sequence encoded as a 64-bit integer.
func check_seed(seed int64, rand_source *rand.Rand) int64{
	rand_source.Seed(seed)
	var seq int64 = 0

	for i := 0; i < SEQ_LEN; i++ {
		val := rand_source.Intn(RANDN)
		// Set the i'th bit
		if val==1 {
			seq |= (1 << uint(i))
		}
	}

	return seq


}

func insertSeedMap(c *mgo.Collection, seqMap map[int64]map[int64]bool){
	// Save a map which maps sequence ints to sets
	// of seeds which are represented by maps of int->bool.
	bulk := c.Bulk()
	bulk.Unordered()
	for seqInt, seedMap := range seqMap {
	    for seed, _ := range seedMap{
		    // match := bson.M{"_id" : seqInt}
		    // update := bson.M{"$addToSet": bson.M{"seeds": seed}}

		    // bulk.Upsert(match, update)

		   	d := bson.M{"seqInt": seqInt, "seed": seed}
		    bulk.Insert(d)

	    }
	}
	r, err := bulk.Run()
	if err != nil {
    	fmt.Println(err)
    	    	fmt.Println(r)

    }
}


// Search for a seed in a specified range that generates the goal sequence. Sends the seed value
// on 'found_seed_ch' when it finds it.
func seed_search(part_id int64, lo int64, hi int64, done_ch chan int64) {
	var batchSize int64 = 1000
	seeds := make(map[int64]map[int64]bool)

	// Connect to database.
	session, err := mgo.Dial(dbstring)
	session.SetSocketTimeout(24 * time.Hour)
    if err != nil {
    	panic(err)
    }
    c := session.DB(dbName).C(collName)

	// Create random generator source.
	rand_source := rand.New(rand.NewSource(0))

	var currSeed int64
	for currSeed = lo; currSeed < hi; currSeed++ {
		seqInt := check_seed(currSeed, rand_source)

		// Save to map
		_, ok := seeds[seqInt]
		if !ok {
    		seedSet := make(map[int64]bool)
        	seedSet[currSeed] = true
    		seeds[seqInt] = seedSet
		} else{
			seeds[seqInt][currSeed] = true
		}

		// Print progress
		if (currSeed % (batchSize*100))==0 {
			progress := float64(currSeed-lo)/float64(hi-lo)*100
			fmt.Printf("(Partition %d) Currently checking seed %d [%f %%]\n", part_id, currSeed, progress)
		}

		// Insert batch.
		if (currSeed % batchSize)==0 {
			insertSeedMap(c, seeds)
			// Clear the map.
			seeds = make(map[int64]map[int64]bool)
		}
	}

	// Last batch.
	insertSeedMap(c, seeds)

	session.Close()

	// Signal 'done' channel when you finish.
	done_ch <- 1

}

func main() {
	runtime.GOMAXPROCS(10)

	args := os.Args[1:]

	var max int64 = (1<<31) - 1

	// var max int64 = 1000000

	// Number of partitions to break the seed space up into.
	i, _ := strconv.ParseInt(args[0], 10, 64)
	var num_partitions int64 = i
	var partition_size int64 = max/num_partitions

	//Channel that will receive the found seed value.
	done_ch := make(chan int64)

	var start = time.Now().Unix()
	var p int64
	for p = 0; p < num_partitions; p++ {
		lo := partition_size * p
		hi := partition_size * (p+1)
		fmt.Printf("Partition %d (%d, %d) \n", p, lo, hi)
		go seed_search(p, lo, hi, done_ch)
	}

	// Wait until all routines finish.
	for i := int64(0); i < num_partitions; i++ {
		<-done_ch
	}

	var end = time.Now().Unix()
	fmt.Printf("Finished seed generation in %d seconds.", end-start)
}
