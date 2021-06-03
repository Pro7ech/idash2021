package main

import(
	"os"
	"fmt"
	"time"
	"encoding/binary"
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/client"
	"github.com/ldsec/idash21_Task2/server"
)
func main(){

	nbGenomes := 2000

	// Key generation
	var err error

	time1 := time.Now()

	var params *ckks.Parameters
	if params, err = ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}}); err != nil {
		panic(err)
	}

	kgen := ckks.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyGaussian()

	fmt.Println()
	fmt.Printf("Key generation done : (%s)\n", time.Since(time1))
	lib.PrintMemUsage()
	fmt.Println()

	// Marshal SecretKey
	var fwSk *os.File
	var b []byte

	if fwSk, err = os.Create(lib.KeysPath); err != nil {
		panic(err)
	}
	defer fwSk.Close()

	if b, err = sk.MarshalBinary(); err != nil {
		panic(err)
	}

	fwSk.Write(b)

	// Pre-processing & Encryption
	client := client.NewClient()

	client.ProcessAndEncrypt("model/X.binary", nbGenomes)

	// Prediction
	server := server.NewServer()

	nbBatches := int(binary.LittleEndian.Uint64(lib.FileToByteBuffer(lib.NbBatchToPredict)))

	time1 = time.Now()
	for i := 0; i < nbBatches; i++ {
		server.PredictBatch(i)
	}
	
	fmt.Printf("Prediction done : %s\n", time.Since(time1))
	lib.PrintMemUsage()
	fmt.Println()
	

	// Decryption
	decryptor := client.NewDecryptor()

	predictions := [][]float64{}

	time1 = time.Now()
	for i := 0; i < nbBatches; i++ {
		ciphertexts := lib.UnmarshalBatch32(lib.EncryptedBatchPredIndexPath(i))
		predictions = append(predictions, decryptor.DecryptBatchTranspose(ciphertexts)...)
	}

	
	fmt.Printf("Decryption done : %s\n", time.Since(time1))
	lib.PrintMemUsage()
	fmt.Println()

	predictions = predictions[:nbGenomes]

	Ybyte := lib.FileToByteBuffer("model/Y.binary")

	acc := 0
	for i, pred := range predictions{
		idx := 0
		max := -10.0
		for j,v := range pred{
			if v > max{
				idx = j
				max = v
			}
		}

		if idx != int(Ybyte[i]){
			acc++
		}
	}

	fmt.Printf("Accuracy : %f (%d errors)\n", 1-float64(acc)/float64(nbGenomes), acc)
}