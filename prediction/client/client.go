package client

import (
	"encoding/binary"
	"github.com/ldsec/idash21_Task2/prediction/lib"
	"github.com/ldsec/lattigo/v2/ckks"
	"log"
	"math"
	"os"
	"sync"
)

type Client struct {
	params *ckks.Parameters
	sk     *ckks.SecretKey
}

func NewClient() (c *Client) {
	var err error
	c = new(Client)
	// Scheme parameters
	if c.params, err = ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi: lib.Q, Pi: []uint64{}}); err != nil {
		log.Fatal(err)
	}
	c.params.SetScale(lib.HashScale)

	// Reads secretkey
	buffsk := lib.FileToByteBuffer(lib.KeysPath)

	// Allocates secretkey
	c.sk = new(ckks.SecretKey)
	if err = c.sk.UnmarshalBinary(buffsk); err != nil {
		log.Fatal(err)
	}

	return
}

func (c *Client) ProcessAndEncrypt(path string) {

	// Encryptor
	encryptor := c.NewEncryptor(lib.NbGoRoutines)

	processedGenomesBytes := lib.FileToByteBuffer(path)

	// Reads the pre-processed genomes
	nbGenomes := int(binary.LittleEndian.Uint64(processedGenomesBytes[:8]))
	processedGenomesBytes = processedGenomesBytes[8:]
	hashes := make([][]float64, nbGenomes)
	for i := range hashes {
		hashes[i] = make([]float64, lib.HashSize)
		for j := range hashes[i] {
			hashes[i][j] = math.Float64frombits(binary.LittleEndian.Uint64(processedGenomesBytes[j<<3 : (j+1)<<3]))
		}
		processedGenomesBytes = processedGenomesBytes[lib.HashSize<<3:]
	}

	// Transpose the matrix of pre-processed genomes
	//
	// 	   Hashes 			     # Genomes
	// 	  ________		       _______________
	// G |dab9072a... 		  |d  2  1  5
	// e |				    H |a  4  8  b
	// n |243527b4...		a |b  3  0  b
	// o | 			   ->	s |9  5  1  2  ...
	// m |18014d82...		h |0  2  4  8
	// e |			    	e |7  7  d  2
	// s |5bb282af...		s |2  b  8  a
	//	 |				      |a  4  2  f

	hashTransposed := make([][]float64, lib.HashSize)
	for i := range hashTransposed {
		hashTransposed[i] = make([]float64, nbGenomes)
	}

	for i := 0; i < nbGenomes; i++ {
		for j := 0; j < lib.HashSize; j++ {
			hashTransposed[j][i] = hashes[i][j]
		}
	}

	//*************************** HASHES ENCRYPTION *******************************

	// We need to encrypt a HashSize x nbGenomes matrix where the i-th row of the
	// matrix is the i-th coefficient of the hash of each genome

	// We encrypt batches of N hashes, each i-th coefficient of the N hashes
	// being stored in its own ciphertext, hence hashSize ciphertexts are needed
	// per batch of H hashes
	// Each batch is encrypted in a different file

	// Number of batches
	nbBatches := int(math.Ceil(float64(nbGenomes) / float64(c.params.N())))

	// Saves how many batches are encrypted
	var fw *os.File
	var err error
	// Creates the files containing the compressed ciphertexts
	if fw, err = os.Create(lib.NbBatchToPredict); err != nil {
		panic(err)
	}
	defer fw.Close()

	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(nbBatches))
	fw.Write(buff)
	binary.LittleEndian.PutUint64(buff, uint64(nbGenomes))
	fw.Write(buff)

	// Number of ciphertext per Go routine
	nbCipherPerGoRoutine := int(math.Ceil(float64(lib.HashSize) / float64(lib.NbGoRoutines)))

	ciphertexts := make([]*ckks.Ciphertext, lib.HashSize)

	// Number of batch
	for i := 0; i < nbBatches; i++ {

		encryptor.Seed()

		var wg sync.WaitGroup
		wg.Add(lib.NbGoRoutines)
		for g := 0; g < lib.NbGoRoutines; g++ {

			start := g * nbCipherPerGoRoutine
			end := (g + 1) * nbCipherPerGoRoutine

			if g == lib.NbGoRoutines-1 {
				end = lib.HashSize
			}

			go func(worker, startHash, endHash int) {

				startGenome := i * int(c.params.N())
				endGenome := (i + 1) * int(c.params.N())
				if endGenome > nbGenomes {
					endGenome = nbGenomes
				}

				tmp := encryptor.Encrypt(worker, startGenome, endGenome, hashTransposed[startHash:endHash])

				for j := startHash; j < endHash; j++ {
					ciphertexts[j] = tmp[j-startHash]
				}

				wg.Done()
			}(g, start, end)
		}
		wg.Wait()

		lib.MarshalBatchSeeded32(lib.EncryptedBatchIndexPath(i), ciphertexts, encryptor.GetSeeds())
	}
}
