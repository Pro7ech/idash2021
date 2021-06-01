package prediction

import(
	"os"
	"fmt"
	"math"
	"time"
	"testing"
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/params"
	"github.com/ldsec/lattigo/v2/ckks"
)

func TestPredictor(t *testing.T){
	schemeParams, _ := ckks.NewParametersFromModuli(params.LogN, &ckks.Moduli{Qi:params.Q, Pi:[]uint64{}})
	predictor := NewPredictor(schemeParams)
	predictor.LoadModel()
	//predictor.PrintModel()

	kgen := ckks.NewKeyGenerator(schemeParams)
	sk := kgen.GenSecretKeyGaussian()
	encoder := ckks.NewEncoder(schemeParams)
	encryptor := ckks.NewEncryptorFromSk(schemeParams, sk)
	decryptor := ckks.NewDecryptor(schemeParams, sk)

	var err error
	var file *os.File
	if file, err = os.Open("../data/X_CGR_DCT"); err != nil{
        panic(err)
    }
    defer file.Close()

    fileInfo, err := file.Stat()
	buff := make([]byte, fileInfo.Size())
	if _, err = file.Read(buff); err != nil {
		panic(err)
	}

	nbHashes := params.NbSamples

	// list of CGR-DCTII hashes 
	hashes := make([][]float64, nbHashes)
    for i := range hashes{
    	tmp := make([]float64, params.HashSize)
    	b := buff[(i*params.HashSize)<<3:(i*params.HashSize+1)<<3]
    	for j := range tmp{
    		tmp[j] = math.Float64frombits(binary.LittleEndian.Uint64(b[j<<3:(j+1)<<3]))
    	}
    	hashes[i] = tmp 
    }

    nbCiphertexts := len(hashes)/int(schemeParams.N())+1

	ciphertexts := make([][]*ckks.Ciphertext, nbCiphertexts)
	plaintext := ckks.NewPlaintext(schemeParams, 0, params.HashScale)
	// For each ciphertext containing N j-th coefficient of a hash
	for i := range ciphertexts{
		ciphertexts[i] = make([]*ckks.Ciphertext, params.HashSize)

		start := i*int(schemeParams.N())
		end := (i+1)*int(schemeParams.N())
		if end > len(hashes){
			end = len(hashes)
		}

		hashSplit := hashes[start:end]

		// For each coefficient of the first i*N hashes

		for j := range ciphertexts[i]{

			values := make([]float64, schemeParams.N())
			for k := 0; k<len(hashSplit); k++{
				values[k] = hashSplit[k][j]
			}

			encoder.EncodeCoeffs(values, plaintext)

			ciphertexts[i][j] = encryptor.EncryptNew(plaintext)
		}
	}

	predictions := make([][]float64, nbHashes)
	for i := range predictions{
		predictions[i] = make([]float64, 4)
	}

	start := time.Now()
	for i := 0; i < params.NbStrains; i++{
		for j := range ciphertexts{
			res := ckks.NewCiphertext(schemeParams, 1, 0, params.HashScale*params.ModelScale)
			predictor.DotProduct(ciphertexts[j], i, res)
			valuesTest := encoder.DecodeCoeffs(decryptor.DecryptNew(res)) 
			
			idx := j*int(schemeParams.N())
			maxN := int(schemeParams.N())
			if j == len(ciphertexts)-1{
				maxN = nbHashes%maxN
			}
			for k := range valuesTest[:maxN]{
				predictions[k+idx][i] = valuesTest[k]
			}
		}
	}
	fmt.Printf("Done : %s\n", time.Since(start))

	acc := 0
	for i := range predictions{
		SoftMax(predictions[i])

		idx := MaxIndex(predictions[i])

		if idx != i/2000{
			acc++
		}
	}

	fmt.Println(1-float64(acc)/8000)
}

func BenchmarkPredictor(b *testing.B){
	schemeParams, _ := ckks.NewParametersFromModuli(params.LogN, &ckks.Moduli{Qi:params.Q, Pi:[]uint64{}})
	predictor := NewPredictor(schemeParams)
	predictor.LoadModel()

	kgen := ckks.NewKeyGenerator(schemeParams)
	sk := kgen.GenSecretKeyGaussian()
	encoder := ckks.NewEncoder(schemeParams)
	encryptor := ckks.NewEncryptorFromSk(schemeParams, sk)

	values := make([]float64, schemeParams.N())
	for i := range values{
		values[i] = 1.0
	}

	plaintext := ckks.NewPlaintext(schemeParams, 0, params.HashScale)
	encoder.EncodeCoeffs(values, plaintext)

	ciphertexts := make([]*ckks.Ciphertext, params.HashSqrtSize*params.HashSqrtSize)

	for i := range ciphertexts{
		ciphertexts[i] = encryptor.EncryptNew(plaintext)
	}

	res := ckks.NewCiphertext(schemeParams, 1, 0, params.HashScale*params.ModelScale)

	b.Run("DotProduct_(1x256)x(256x4)", func(b *testing.B) {

		for i := 0; i < b.N; i++{
			predictor.DotProduct(ciphertexts, 0, res)
			predictor.DotProduct(ciphertexts, 1, res)
			predictor.DotProduct(ciphertexts, 2, res)
			predictor.DotProduct(ciphertexts, 3, res)
		}
	})
}