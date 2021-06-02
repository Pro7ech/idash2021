package client


import(
	"fmt"
	"testing"
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/idash21_Task2/lib"
)

func TestNewClient(t *testing.T){
	client := NewClient()
	client.ProcessAndEncrypt(4, 2000)

	ciphertexts := lib.UnmarshalBatchSeeded32("../"+lib.EncryptedBatchIndexPath(1))

	decryptor := ckks.NewDecryptor(client.params, client.sk)
	encoder := ckks.NewEncoder(client.params)

	values := encoder.DecodeCoeffs(decryptor.DecryptNew(ciphertexts[1]))

	fmt.Println(values[:8])
	fmt.Println(values[len(values)-8:len(values)-1])
}

