package main

import(
	"os"
	"math"
	"strconv"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/client"
)

func main(){

	args := os.Args[1:]
	if len(args) == 0 {
		panic("NEED NBGENOMES")
	}

	nbGenomes, _ := strconv.Atoi(args[0])

	// Creates a new client
	// Expect a secret-key in key/
	client := client.NewClient()

	decryptor := client.NewDecryptor()

	nbBatches := int(math.Ceil(float64(nbGenomes)/float64(int(1<<lib.LogN))))

	predictions := [][]float64{}

	for i := 0; i < nbBatches; i++{
		ciphertexts := lib.UnmarshalBatch32(lib.EncryptedBatchPredIndexPath(i))
		predictions = append(predictions, decryptor.DecryptBatchTranspose(ciphertexts)...)
	}

	predictions = predictions[:nbGenomes]
}