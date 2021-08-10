package server

import (
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/idash21_Task2/prediction/predictor"
	"github.com/ldsec/lattigo/v2/ckks"
	"sync"
)

type Server struct {
	params    *ckks.Parameters
	predictor *predictor.Predictor
}

func NewServer() (server *Server) {
	params, _ := ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}})
	predictor := predictor.NewPredictor(params)
	predictor.LoadModel(lib.ModelPath)
	//predictor.PrintModel()
	return &Server{params: params, predictor: predictor}
}

func (s *Server) PredictBatch(batchIndex int) {
	// Unmarchal batch to predict
	ciphertexts := lib.UnmarshalBatchSeeded32(lib.EncryptedBatchIndexPath(batchIndex))

	// Allocates results
	pred := make([]*ckks.Ciphertext, lib.NbStrains)
	for i := range pred {
		pred[i] = ckks.NewCiphertext(s.params, 1, 0, lib.HashScale*lib.ModelScale)
	}

	var wg sync.WaitGroup
	wg.Add(lib.NbStrains)
	for i := 0; i < lib.NbStrains; i++ {

		go func(worker int, pred *ckks.Ciphertext) {
			s.predictor.DotProduct(ciphertexts, worker, pred)
			wg.Done()
		}(i, pred[i])

	}
	wg.Wait()

	// Marchal prediction
	lib.MarshalBatch32(lib.EncryptedBatchPredIndexPath(batchIndex), pred)
}
