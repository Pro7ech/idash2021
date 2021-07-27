package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/preprocessing"
	"github.com/ldsec/lattigo/v2/ckks"
	"log"
	"math"
	"os"
	"sync"
	"time"
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

func (c *Client) ProcessAndEncrypt(path string, nbGenomes int) {

	//*************************** GENOMES PRE-PROCESSING *******************************

	// Chaos Game Representation + 2D Discret Cosine II hasher
	hasher := preprocessing.NewDCTHasher(1, lib.Window, lib.HashSqrtSize, lib.Normalizer)

	// Encryptor
	encryptor := c.NewEncryptor(lib.NbGoRoutines)

	// Allocate genomes hash list
	hashes := make([][]float64, nbGenomes)
	for i := range hashes {
		hashes[i] = make([]float64, lib.HashSize)
	}

	// Reads the genomes

	var err error
	file, err := os.Open(lib.GenomeDataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	i := 0
	data := make([]string, lib.NbGoRoutines)
	remain := nbGenomes % lib.NbGoRoutines
	time1 := time.Now()
	for scanner.Scan() {

		if i%20 == 0 && (i>>1) < nbGenomes {
			fmt.Printf("\rProcessing %4d Genomes :%3d%%", nbGenomes, int(100*(float64(i>>1)/float64(nbGenomes))))
		}

		// Expects :
		// Even indexes = genome ID
		// Odd indexes = genome
		if i&1 == 1 && (i>>1) < nbGenomes {

			// Assigns a genome to the data list
			data[(i>>1)%lib.NbGoRoutines] = scanner.Text()

			// Once data list is filled process the genomes
			// or if reached the last genome, processes the data list
			if (i>>1)%lib.NbGoRoutines == lib.NbGoRoutines-1 || (i>>1) == nbGenomes-1 {

				nbToProcess := lib.NbGoRoutines
				if (i>>1) == nbGenomes-1 && remain != 0 {
					nbToProcess = remain
				}

				var wg sync.WaitGroup
				wg.Add(nbToProcess)
				for g := 0; g < nbToProcess; g++ {
					go func(worker int, genome string) {
						if (i>>1)+worker-nbToProcess+1 < nbGenomes {
							hasher.Hash(worker, genome) // CGR + 2D DCTII hashing
							copy(hashes[(i>>1)+worker-nbToProcess+1], hasher.GetHash(worker))
						}
						wg.Done()
					}(g, data[g])
				}
				wg.Wait()
			}
		}
		i++
	}

	// Transpose the hash matrix
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

	fmt.Printf("\rProcessing %4d Genomes : %3d%% (%s)\n", nbGenomes, 100, time.Since(time1))
	lib.PrintMemUsage()
	fmt.Println()

	//fmt.Println(hashes[0])

	//fmt.Println(hashTransposed[0][len(hashTransposed[0])-1])
	//fmt.Println(hashes[len(hashes)-1][0])

	//*************************** HASHES ENCRYPTION *******************************

	// We need to encrypt a HashSize x nbGenomes matrix where the i-th row of the
	// matrix is the i-th coefficient of the hash of each genome

	time1 = time.Now()

	// We encrypt batches of N hashes, each i-th coefficient of the N hashes
	// being stored in its own ciphertext, hence hashSize ciphertexts are needed
	// per batch of H hashes
	// Each batch is encrypted in a different file

	// Number of batches
	nbBatches := int(math.Ceil(float64(nbGenomes) / float64(c.params.N())))

	// Saves how many batches are encrypted
	var fw *os.File
	// Creates the files containing the compressed ciphertexts
	if fw, err = os.Create(lib.NbBatchToPredict); err != nil {
		panic(err)
	}
	defer fw.Close()

	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(nbBatches))
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

			fmt.Printf("\rEncrypting %4d Hashes  : %3d%%", nbGenomes, int(100*float64(i*lib.NbGoRoutines+g)/float64(nbBatches*lib.NbGoRoutines)))

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

	fmt.Printf("\rEncrypting %4d Hashes  : %3d%% (%s)\n", nbGenomes, 100, time.Since(time1))
	lib.PrintMemUsage()
	fmt.Println()

}
