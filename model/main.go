package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/preprocessing"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

func main() {

	// Preprocessing for model training

	strainMap := lib.StrainsMap
	nbSamplesStrain := lib.NbSamplesPerStrain
	nbSamples := lib.NbSamples
	hashsqrtsize := lib.HashSqrtSize
	window := lib.Window
	nbGo := 4

	var err error
	file, err := os.Open("./data/Challenge.fa")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	buff := make([]byte, hashsqrtsize*hashsqrtsize*8)

	// Creates the files containing the processed samples
	var fw *os.File
	if fw, err = os.Create("X_CGR_DCT.binary"); err != nil {
		panic(err)
	}

	hasher := preprocessing.NewDCTHasher(nbGo, window, hashsqrtsize)

	if fw, err = os.OpenFile("X_CGR_DCT.binary", os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		panic(err)
	}

	defer file.Close()

	start := time.Now()

	i := 0
	data := make([]string, nbGo)
	for scanner.Scan() {

		if i%200 == 1 {
			fmt.Printf("%-7s : %4d/%d\n", strainMap[i/(2*nbSamplesStrain)], i>>1, nbSamples)
		}

		if i&1 == 1 {

			data[(i>>1)%nbGo] = scanner.Text()

			if (i>>1)%nbGo == nbGo-1 {

				var wg sync.WaitGroup
				wg.Add(nbGo)
				for g := 0; g < nbGo; g++ {
					go func(worker int, strain string) {
						hasher.Hash(worker, strain)
						wg.Done()
					}(g, data[g])
				}
				wg.Wait()

				for g := 0; g < nbGo; g++ {
					hash := hasher.GetHash(g)
					for i := range hash {
						binary.LittleEndian.PutUint64(buff[i<<3:(i+1)<<3], math.Float64bits(hash[i]))
					}

					fw.Write(buff)
				}
			}
		}
		i++
	}

	fmt.Println("Done :", time.Since(start))
}
