package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/idash21_Task2/prediction/preprocessing"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

func main() {

	var err error

	// Preprocessing for model training

	nbSamples := 8000
	hashsqrtsize := lib.HashSqrtSize
	window := lib.Window         // SEE **** WARNING *****
	normalizer := lib.Normalizer // applies x -> x^normalizer to the FCGR probability matrix
	nbGo := 4

	fmt.Printf("Pre-processing\n")
	fmt.Printf("Samples : %d\n", nbSamples)
	fmt.Printf("Window : %d\n", window)
	fmt.Printf("Normalizer : x^%f\n", normalizer)
	fmt.Printf("Hashs Size : %d\n", hashsqrtsize*hashsqrtsize)

	// Writes the processing parameters for the .py file training
	var fwParams *os.File
	if fwParams, err = os.Create("./params.binary"); err != nil {
		panic(err)
	}

	fwParams.Write([]byte{uint8(hashsqrtsize)})
	fwParams.Close()

	// ****** WARNING *****

	// If choosing this hasher, then the FCGR matrix will be of size  **** 4^window ****
	// For this option, window can be either even or odd
	hasher := preprocessing.NewDCTHasher(nbGo, window, hashsqrtsize, normalizer)

	// If choosing this hasher then the FCGR matrix will of of size **** 2^(window+2) ****
	// For this case window must ONLY be even
	//hasher := preprocessing.NewDCTHasherV2(nbGo, window, hashsqrtsize, normalizer)

	file, err := os.Open("./Challenge.fa")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	buffX := make([]byte, hashsqrtsize*hashsqrtsize*8)
	buffY := make([]byte, nbGo)

	// Creates the files containing the processed samples
	var fwX, fwY *os.File
	if fwX, err = os.Create("./X.binary"); err != nil {
		panic(err)
	}

	if fwY, err = os.Create("./Y.binary"); err != nil {
		panic(err)
	}

	var fwXCSV, fwYCSV *os.File
	if fwXCSV, err = os.Create("./X.csv"); err != nil {
		panic(err)
	}

	if fwYCSV, err = os.Create("./Y.csv"); err != nil {
		panic(err)
	}

	wX := csv.NewWriter(fwXCSV)
	defer wX.Flush()

	wY := csv.NewWriter(fwYCSV)
	defer wY.Flush()

	start := time.Now()

	i := 0
	dataX := make([]string, nbGo)
	dataY := make([]string, nbGo)
	dataCSV := make([]string, 256)
	for scanner.Scan() {

		if i&1 == 0 {
			dataY[(i>>1)%nbGo] = scanner.Text()
			if i%200 == 0 {
				fmt.Printf("\rProcessing samples: %4d/%d", i>>1, nbSamples)
			}
		}

		if i&1 == 1 {

			dataX[(i>>1)%nbGo] = scanner.Text()

			if (i>>1)%nbGo == nbGo-1 {

				var wg sync.WaitGroup
				wg.Add(nbGo)
				for g := 0; g < nbGo; g++ {
					go func(worker int, strain string) {
						hasher.MapCGR(worker, strain)
						hasher.DCTII(worker)
						hasher.Finalize(worker)
						wg.Done()
					}(g, dataX[g])
				}
				wg.Wait()

				for g := 0; g < nbGo; g++ {
					hash := hasher.GetHash(g)
					for i := range hash {
						binary.LittleEndian.PutUint64(buffX[i<<3:(i+1)<<3], math.Float64bits(hash[i]))

						dataCSV[i] = fmt.Sprintf("%f", hash[i])
					}

					buffY[g] = uint8(MatchStrainNameToLabel(dataY[g]))

					fwX.Write(buffX)

					wX.Write(dataCSV)
					wY.Write([]string{fmt.Sprintf("%d", int(buffY[g]))})
				}

				fwY.Write(buffY)
			}

			if (i >> 1) == nbSamples {
				break
			}
		}
		i++
	}

	fwX.Close()
	fwY.Close()

	fmt.Printf("\rProcessing samples: %4d/%d (%s)\n", nbSamples, nbSamples, time.Since(start))
}

func MatchStrainNameToLabel(substring string) (label int) {

	strain := ""
	for _, c := range substring[1:] {
		if string(c) != "_" {
			strain += string(c)
		} else {
			break
		}
	}

	return lib.StrainsMap[strain]
}
