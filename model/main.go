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

	buffX := make([]byte, hashsqrtsize*hashsqrtsize*8)
    buffY := make([]byte, nbGo)

	// Creates the files containing the processed samples
	var fwX, fwY *os.File
	if fwX, err = os.Create("./model/X.binary"); err != nil {
		panic(err)
	}

    if fwY, err = os.Create("./model/Y.binary"); err != nil{
        panic(err)
    }

	hasher := preprocessing.NewDCTHasher(nbGo, window, hashsqrtsize)

	start := time.Now()

	i := 0
	dataX := make([]string, nbGo)
    dataY := make([]string, nbGo)
	for scanner.Scan() {

        if i&1 == 0{
            dataY[(i>>1)%nbGo] = scanner.Text()
            if i%200 == 0{
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
						hasher.Hash(worker, strain)
						wg.Done()
					}(g, dataX[g])
				}
				wg.Wait()

				for g := 0; g < nbGo; g++ {
					hash := hasher.GetHash(g)
					for i := range hash {
						binary.LittleEndian.PutUint64(buffX[i<<3:(i+1)<<3], math.Float64bits(hash[i]))
					}

                    buffY[g] = uint8(MatchStrainNameToLabel(dataY[g]))

					fwX.Write(buffX)
				}

                fwY.Write(buffY)
			}
		}
		i++
	}

    fwX.Close()
    fwY.Close()

	fmt.Printf("\rProcessing samples: %4d/%d (%s)", nbSamples, nbSamples, time.Since(start))
}


func MatchStrainNameToLabel(substring string) (label int){

    strain := ""
    for _, c := range substring[1:]{
        if string(c) != "_"{
            strain += string(c)
        }else{
            break
        }
    }

    return lib.StrainsMap[strain]
}
