package main

import (
	"github.com/ldsec/idash21_Task2/prediction/client"
)

func main() {

	// Creates a new client
	// Expect a secret-key in key/
	client := client.NewClient()

	// 1) Opens the file containing genomes
	// 		- Except a file where even indexes are the genome ID
	// 		  and odd indexes are the genome data
	// 2) Processes the first xx genomes
	// 3) Encrypts the processed genomes by batches of 1<<lib.LogN
	// 4) Saves each batch in a separate file in temp/enc_client_batch_{i}.binary
	client.ProcessAndEncrypt("temps/preprocessed.binary")
}
