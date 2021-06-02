package client

import (
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/lattigo/v2/ckks"
)

// Decryptor is a struct storing the necessary object to decrypt and decode ciphertexts.
type Decryptor struct {
	decryptor ckks.Decryptor
	encoder   ckks.Encoder
	plaintext *ckks.Plaintext
}

// NewDecryptor creates a new Decryptor.
func (c *Client) NewDecryptor() (decryptor *Decryptor) {
	decryptor = new(Decryptor)
	decryptor.decryptor = ckks.NewDecryptor(c.params, c.sk)
	decryptor.encoder = ckks.NewEncoder(c.params)
	decryptor.plaintext = ckks.NewPlaintext(c.params, 0, 0)
	return
}


func (d *Decryptor) DecryptBatch(ciphertexts []*ckks.Ciphertext) (pred [][]float64){
	pred = make([][]float64, len(ciphertexts))
	for i := range pred{
		pred[i] = make([]float64, 1<<lib.LogN)
	}
	for i := range ciphertexts{
		d.decryptor.Decrypt(ciphertexts[i], d.plaintext)

		v := d.encoder.DecodeCoeffs(d.plaintext)

		tmp := pred[i]
		for j := range v{
			tmp[j] = v[j]
		}
	}
	
	return
}

func (d *Decryptor) DecryptBatchTranspose(ciphertexts []*ckks.Ciphertext) (pred [][]float64){
	pred = make([][]float64, 1<<lib.LogN)
	for i := range pred{
		pred[i] = make([]float64, len(ciphertexts))
	}
	for i := range ciphertexts{
		d.decryptor.Decrypt(ciphertexts[i], d.plaintext)

		v := d.encoder.DecodeCoeffs(d.plaintext)

		for j := range pred{
			pred[j][i] = v[j]
		}
	}

	return
}