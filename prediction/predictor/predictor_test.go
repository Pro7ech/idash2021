package predictor

import (
	"encoding/binary"
	"fmt"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/lattigo/v2/ckks"
	"math"
	"os"
	"testing"
	"time"
)

func TestPredictor(t *testing.T) {
	params, _ := ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}})
	predictor := NewPredictor(params)
	predictor.LoadModel("../" + lib.ModelPath)
	//predictor.PrintModel()

	kgen := ckks.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyGaussian()
	encoder := ckks.NewEncoder(params)
	encryptor := ckks.NewEncryptorFromSk(params, sk)
	decryptor := ckks.NewDecryptor(params, sk)

	var err error
	var file *os.File
	if file, err = os.Open("../data/X_CGR_DCT"); err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	buff := make([]byte, fileInfo.Size())
	if _, err = file.Read(buff); err != nil {
		panic(err)
	}

	nbHashes := lib.NbSamples

	// list of CGR-DCTII hashes
	hashes := make([][]float64, nbHashes)
	for i := range hashes {
		tmp := make([]float64, lib.HashSize)
		b := buff[(i*lib.HashSize)<<3 : (i*lib.HashSize+1)<<3]
		for j := range tmp {
			tmp[j] = math.Float64frombits(binary.LittleEndian.Uint64(b[j<<3 : (j+1)<<3]))
		}
		hashes[i] = tmp
	}

	nbCiphertexts := len(hashes)/int(params.N()) + 1

	start := time.Now()
	ciphertexts := make([][]*ckks.Ciphertext, nbCiphertexts)
	plaintext := ckks.NewPlaintext(params, 0, lib.HashScale)
	// For each ciphertext containing N j-th coefficient of a hash
	for i := range ciphertexts {
		ciphertexts[i] = make([]*ckks.Ciphertext, lib.HashSize)

		start := i * int(params.N())
		end := (i + 1) * int(params.N())
		if end > len(hashes) {
			end = len(hashes)
		}

		hashSplit := hashes[start:end]

		// For each coefficient of the first i*N hashes

		for j := range ciphertexts[i] {

			values := make([]float64, params.N())
			for k := 0; k < len(hashSplit); k++ {
				values[k] = hashSplit[k][j]
			}

			encoder.EncodeCoeffs(values, plaintext)

			ciphertexts[i][j] = encryptor.EncryptNew(plaintext)
		}
	}
	fmt.Printf("Done : %s\n", time.Since(start))

	predictions := make([][]float64, nbHashes)
	for i := range predictions {
		predictions[i] = make([]float64, 4)
	}

	start = time.Now()
	for i := 0; i < lib.NbStrains; i++ {
		for j := range ciphertexts {
			res := ckks.NewCiphertext(params, 1, 0, lib.HashScale*lib.ModelScale)
			predictor.DotProduct(ciphertexts[j], i, res)
			valuesTest := encoder.DecodeCoeffs(decryptor.DecryptNew(res))

			idx := j * int(params.N())
			maxN := int(params.N())
			if j == len(ciphertexts)-1 {
				maxN = nbHashes % maxN
			}
			for k := range valuesTest[:maxN] {
				predictions[k+idx][i] = valuesTest[k]
			}
		}
	}
	fmt.Printf("Done : %s\n", time.Since(start))

	acc := 0
	for i := range predictions {
		SoftMax(predictions[i])

		idx := MaxIndex(predictions[i])

		if idx != i/2000 {
			acc++
		}
	}

	fmt.Println(1 - float64(acc)/8000)
}

func BenchmarkPredictor(b *testing.B) {
	params, _ := ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}})
	predictor := NewPredictor(params)
	predictor.LoadModel("../" + lib.ModelPath)

	kgen := ckks.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyGaussian()
	encoder := ckks.NewEncoder(params)
	encryptor := ckks.NewEncryptorFromSk(params, sk)

	values := make([]float64, params.N())
	for i := range values {
		values[i] = 1.0
	}

	plaintext := ckks.NewPlaintext(params, 0, lib.HashScale)
	encoder.EncodeCoeffs(values, plaintext)

	ciphertexts := make([]*ckks.Ciphertext, lib.HashSqrtSize*lib.HashSqrtSize)

	for i := range ciphertexts {
		ciphertexts[i] = encryptor.EncryptNew(plaintext)
	}

	res := ckks.NewCiphertext(params, 1, 0, lib.HashScale*lib.ModelScale)

	b.Run("DotProduct_(1x256)x(256x4)", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			predictor.DotProduct(ciphertexts, 0, res)
			predictor.DotProduct(ciphertexts, 1, res)
			predictor.DotProduct(ciphertexts, 2, res)
			predictor.DotProduct(ciphertexts, 3, res)
		}
	})
}
