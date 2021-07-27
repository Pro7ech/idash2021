package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/ldsec/idash21_Task2/client"
	"github.com/ldsec/idash21_Task2/lib"
	"log"
	"math"
	"os"
	"encoding/binary"
)

func main() {

	var err error

	// Creates a new client
	// Expect a secret-key in key/
	client := client.NewClient()

	decryptor := client.NewDecryptor()

	// Reads the number of genomes
	var fr *os.File
	if fr, err = os.Open(lib.NbBatchToPredict); err != nil {
		panic(err)
	}

	buff := make([]byte, 16)
	if _, err := fr.Read(buff); err != nil {
		panic(err)
	}

	nbGenomes := int(binary.LittleEndian.Uint64(buff[8:]))

	nbBatches := int(math.Ceil(float64(nbGenomes) / float64(int(1<<lib.LogN))))

	predictions := [][]float64{}

	for i := 0; i < nbBatches; i++ {
		ciphertexts := lib.UnmarshalBatch32(lib.EncryptedBatchPredIndexPath(i))
		predictions = append(predictions, decryptor.DecryptBatchTranspose(ciphertexts)...)
	}

	predictions = predictions[:nbGenomes]

	resf, err := os.Open(lib.GenomeDataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer resf.Close()
	scanner := bufio.NewScanner(resf)

	predf, err := os.Create("results/prediction.csv")
	defer predf.Close()

	w := csv.NewWriter(predf)
	defer w.Flush()

	var i int
	var data = make([]string, 5)
	for scanner.Scan() {

		if i == nbGenomes {
			break
		}

		if i&1 == 0 {

			pred := predictions[i>>1]

			data[0] = scanner.Text()
			data[1] = fmt.Sprintf("%f", pred[0])
			data[2] = fmt.Sprintf("%f", pred[1])
			data[3] = fmt.Sprintf("%f", pred[2])
			data[4] = fmt.Sprintf("%f", pred[3])

			if err = w.Write(data); err != nil {
				log.Fatal(err)
			}
		}

		i++
	}
}
