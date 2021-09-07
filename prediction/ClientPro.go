package main

import (
	"bufio"
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/idash21_Task2/prediction/preprocessing"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
)

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		panic("NEED NBGENOMES")
	}

	nbGenomes, _ := strconv.Atoi(args[0])

	//*************************** GENOMES PRE-PROCESSING *******************************

	// Chaos Game Representation + 2D Discret Cosine II hasher
	hasher := preprocessing.NewDCTHasher(1, lib.Window, lib.HashSqrtSize, lib.Normalizer)

	// Allocate genomes hash list
	hashes := make([][]float64, nbGenomes)
	for i := range hashes {
		hashes[i] = make([]float64, lib.HashSize)
	}

	// Reads the genomes
	var err error
	file, err := os.Open(lib.GenomeDataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	i := 0
	data := make([]string, 1) // 1x Go Routine
	remain := nbGenomes % 1
	for scanner.Scan() {

		// Expects :
		// Even indexes = genome ID
		// Odd indexes = genome
		if i&1 == 1 && (i>>1) < nbGenomes {

			// Assigns a genome to the data list
			data[(i>>1)%1] = scanner.Text()

			// Once data list is filled process the genomes
			// or if reached the last genome, processes the data list
			if (i>>1)%1 == 0 || (i>>1) == nbGenomes-1 {

				nbToProcess := 1 // 1x Go Routine
				if (i>>1) == nbGenomes-1 && remain != 0 {
					nbToProcess = remain
				}

				var wg sync.WaitGroup
				wg.Add(nbToProcess)
				for g := 0; g < nbToProcess; g++ {
					go func(worker int, genome string) {
						if (i>>1)+worker-nbToProcess+1 < nbGenomes {
							hasher.Hash(worker, genome) // CGR + 2D DCTII hashing
							copy(hashes[(i>>1)+worker-nbToProcess+1], hasher.GetHash(worker))
						}
						wg.Done()
					}(g, data[g])
				}
				wg.Wait()
			}
		}
		i++
	}

	// Writes the pre-processed genomes in the /temps folder
	var fw *os.File
	if fw, err = os.Create("temps/preprocessed.binary"); err != nil {
		panic(err)
	}
	defer fw.Close()

	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(nbGenomes))
	fw.Write(buff)

	buff = make([]byte, lib.HashSize<<3)
	for i := range hashes {
		for j, c := range hashes[i] {
			binary.LittleEndian.PutUint64(buff[j<<3:(j+1)<<3], math.Float64bits(c))
		}
		fw.Write(buff)
	}
}
