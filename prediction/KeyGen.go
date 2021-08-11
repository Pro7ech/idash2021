package main

import (
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/lattigo/v2/ckks"
	"os"
)

func main() {
	var err error

	// Generates CKKS parameters
	var params *ckks.Parameters
	if params, err = ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}}); err != nil {
		panic(err)
	}

	// Generates a Gaussian secret-key
	kgen := ckks.NewKeyGenerator(params)
	sk := kgen.GenSecretKeyGaussian()

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
}
