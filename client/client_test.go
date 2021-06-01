package client


import(
	"testing"
	
)

func TestNewClient(t *testing.T){
	client := NewClient()

	client.ProcessAndEncrypt(4, "../data/Challenge.fa", 2000)
}