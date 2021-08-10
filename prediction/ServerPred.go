package main

import (
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/idash21_Task2/prediction/server"
	"time"
	"fmt"
)

func main() {
	server := server.NewServer()

	// Read the number of batches to predict
	nbBatches := int(binary.LittleEndian.Uint64(lib.FileToByteBuffer(lib.NbBatchToPredict)))

	time1 := time.Now()

	// For each batch (file), predicts the values
	for i := 0; i < nbBatches; i++ {
		server.PredictBatch(i)
	}

	lib.PrintMemUsage()
	time2 := time.Now()
	fmt.Printf("[Prediction] %d batches : done %f s\n", nbBatches, time2.Sub(time1).Seconds())
}
