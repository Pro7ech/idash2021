package main

import(
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/server"
)

func main(){
	server := server.NewServer()

	nbBatches := int(binary.LittleEndian.Uint64(lib.FileToByteBuffer(lib.NbBatchToPredict)))

	for i := 0; i < nbBatches; i++{
		server.PredictBatch(i)
	}
}