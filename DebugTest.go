package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ldsec/idash21_Task2/client"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/server"
	"github.com/ldsec/lattigo/v2/ckks"
	"os"
	"time"
	"strconv"
)

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		panic("NEED NBGENOMES")
	}

	nbGenomes, _ := strconv.Atoi(args[0])

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

	TP := []int{0, 0, 0, 0}
	TN := []int{0, 0, 0, 0}
	FP := []int{0, 0, 0, 0}
	FN := []int{0, 0, 0, 0}
	for i, pred := range predictions {
		idx := 0
		max := -4294967296.0
		for j, v := range pred {
			if v > max {
				idx = j
				max = v
			}
		}

		if idx != int(Ybyte[i]) {
			FP[idx]++
			for i := range FN {
				if i != idx {
					FN[i]++
				}
			}
		} else {
			TP[idx]++
			for i := range TN {
				if i != idx {
					TN[i]++
				}
			}
		}
	}
	fmt.Println("True  positives :", TP)
	fmt.Println("True  Negatives :", TN)
	fmt.Println("False Positives :", FP)
	fmt.Println("False Negatives :", FN)
	fmt.Println()

	macro := 0.0
	microNum := 0.0
	microDen := 0.0
	for i := range TP {
		macroNum := float64(TP[i])
		macroDen := float64(TP[i] + FP[i])
		if macroDen != 0 {
			macro += macroNum / macroDen
		}
		microNum += float64(TP[i])
		microDen += float64(FP[i] + TP[i])
	}

	macro /= float64(4)
	micro := microNum / microDen

	fmt.Printf("Macro AUC: %f\n", macro)
	fmt.Printf("Micro AUC: %f\n", micro)
}
