package prediction

import(
	"testing"
	"github.com/ldsec/idash21_Task2/params"
	"github.com/ldsec/lattigo/v2/ckks"
)

func TestPredictor(t *testing.T){
	schemeParams, _ := ckks.NewParametersFromModuli(params.LogN, &ckks.Moduli{Qi:params.Q, Pi:[]uint64{}})
	predictor := NewPredictor(schemeParams)
	predictor.LoadModel()
	predictor.PrintModel()

	kgen := ckks.NewKeyGenerator(schemeParams)
	sk := kgen.GenSecretKeyGaussian()
	encoder := ckks.NewEncoder(schemeParams)
	encryptors := ckks.NewEncryptorFromSk(schemeParams, sk)

	values := make([]float64, schemeParams.N())
	for i := range values{
		values[i] = 1.0
	}

	plaintext := ckks.NewPlaintext(schemeParams, 0, params.HashScale)
	encoder.EncodeCoeffs(values, plaintext)

	ciphertexts := make([]*ckks.Ciphertext, params.HashSqrtSize*params.HashSqrtSize)

	for i := range ciphertexts{
		ciphertexts[i] = encryptors.EncryptNew(plaintext)
	}

	res := ckks.NewCiphertext(schemeParams, 1, 0, params.HashScale*params.ModelScale)

	predictor.DotProduct(ciphertexts, 0, res)
}