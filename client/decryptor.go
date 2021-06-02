package client

import (
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

func (d *Decryptor) Decrypt() (pred [][]float64){
	return
}