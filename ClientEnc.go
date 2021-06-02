package main

import(
	//"fmt"
	"github.com/ldsec/idash21_Task2/client"
	"github.com/ldsec/idash21_Task2/lib"
)

func main(){

	// Creates a new client
	// Expect a secret-key in key/
	client := client.NewClient()

	// 1) Opens the file containing genomes
	// 		- Except a file where odd indexes are the genome ID 
	// 		  and even indexes are the genome data
	// 2) Processes the first xx genomes
	// 3) Encrypts the processed genomes by batch of 1<<lib.LogN
	// 4) Saves each batch in a separate file in temp/enc_client_batch_{i}.binary
	client.ProcessAndEncrypt(lib.GenomeDataPath, 2000)

	/*
	ciphertexts := lib.UnmarshalBatchSeeded32(lib.EncryptedBatchIndexPath(0))

	decryptor := client.NewDecryptor()

	pred := decryptor.DecryptBatch(ciphertexts)

	for i := range pred{
		fmt.Println(i, pred[i][0])
	}
	ciphertexts := lib.UnmarshalBatchSeeded32("../"+lib.EncryptedBatchIndexPath(1))

	decryptor := ckks.NewDecryptor(client.params, client.sk)
	encoder := ckks.NewEncoder(client.params)

	values := encoder.DecodeCoeffs(decryptor.DecryptNew(ciphertexts[1]))

	fmt.Println(values[:8])
	fmt.Println(values[len(values)-8:len(values)-1])
	*/
}