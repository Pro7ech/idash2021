package main

import (
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/idash21_Task2/prediction/server"
)

func main() {
	server := server.NewServer()

	// Read the number of batches to predict
	nbBatches := int(binary.LittleEndian.Uint64(lib.FileToByteBuffer(lib.NbBatchToPredict)))

	// For each batch (file), predicts the values
	for i := 0; i < nbBatches; i++ {
		server.PredictBatch(i)
	}
}
