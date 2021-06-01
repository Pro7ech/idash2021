package client

import(
	"os"
	"log"
	"sync"
	"math"
	"bufio"
	"fmt"
	"github.com/ldsec/lattigo/v2/ckks"
	//"github.com/ldsec/idash21_Task2/utils"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/idash21_Task2/preprocessing"
)

type Client struct{
	params *ckks.Parameters
	sk *ckks.SecretKey
}

func NewClient() (c *Client){
	var err error
	c = new(Client)
	// Scheme parameters
	if c.params, err = ckks.NewParametersFromModuli(lib.LogN, &ckks.Moduli{Qi:lib.Q, Pi:[]uint64{}}); err != nil{
		log.Fatal(err)
	}
	c.params.SetScale(lib.HashScale)


	/*
	// Reads secretkey
	var buffsk []byte
	if buffsk, err = utils.FileToByteBuffer("Keys/SecretKey.binary"); err != nil{
		log.Fatal(err)
	}

	// Allocates secretkey
	c.sk = new(ckks.SecretKey)
	if err = c.sk.UnmarshalBinary(buffsk); err != nil {
		log.Fatal(err)
	}
	*/

	kgen := ckks.NewKeyGenerator(c.params)
	c.sk = kgen.GenSecretKeyGaussian()

	return
}


func (c *Client) ProcessAndEncrypt(nbGoRoutines int, dataPath string, nbGenomes int){

	var err error 
	file, err := os.Open(dataPath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    hasher := preprocessing.NewDCTHasher(nbGoRoutines, lib.Window, lib.HashSqrtSize)
    encryptor := c.NewEncryptor(nbGoRoutines)

    hashes := make([][]float64, nbGenomes)
    for i := range hashes{
    	hashes[i] = make([]float64, lib.HashSize)
    }

    // Data pre-processing
    i := 0
    data := make([]string, nbGoRoutines)
    remain := nbGenomes%nbGoRoutines
    fmt.Printf("Processing Genomes...\n")
    for scanner.Scan() {

    	//if i%200==0 && (i>>1) < nbGenomes{
    	//	fmt.Printf("% 5d/%d...\n", i>>1, nbGenomes)
    	//}

    	// Even indexes = name
    	// Odd indexes = genome
        if i&1 == 1 && (i>>1) < nbGenomes{

        	// Assigns a genom to the data list
			data[(i>>1)%nbGoRoutines] = scanner.Text() 

			// Once all data list are filled process the genomes
        	if (i>>1)%nbGoRoutines == nbGoRoutines-1 || (i>>1) == nbGenomes-1{

        		nbToProcess := nbGoRoutines
        		if (i>>1) == nbGenomes-1{
        			nbToProcess = remain
        		}

                var wg sync.WaitGroup
                wg.Add(nbToProcess)
			    for g := 0; g < nbToProcess; g++{
					go func(worker int, genome string){
						if (i>>1)+worker < nbGenomes{
						    hasher.Hash(worker, genome)
							copy(hashes[(i>>1)+worker], hasher.GetHash(worker))
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
    hashTransposed := make([][]float64, lib.HashSize)
    for i := range hashTransposed{
    	hashTransposed[i] = make([]float64, nbGenomes)
    }

    for i := 0; i < nbGenomes; i++{
    	for j := 0; j < lib.HashSize; j++{
    		hashTransposed[j][i] = hashes[i][j]
    	}
    }

    // We need to encrypt a HashSize x nbGenomes matrix where the i-th row of the
    // matrix is the i-th coefficient of the hash of each genome

    // Number of ciphertexts per row of the matrix
    nbCiphertextsPerCoefficient := int(math.Ceil(float64(nbGenomes)/float64(c.params.N())))

    // Number of ciphertext per column of the matrix
    coefficientsPerGoRoutine := int(math.Ceil(float64(lib.HashSize)/float64(nbGoRoutines)))

    fmt.Println(nbCiphertextsPerCoefficient)
    fmt.Println(coefficientsPerGoRoutine)

    ciphertexts := make([][]*ckks.Ciphertext, nbCiphertextsPerCoefficient)

    // We encrypt batch of coefficients of eatch row
    for i := 0; i < nbCiphertextsPerCoefficient; i++{

    	ciphertexts[i] = make([]*ckks.Ciphertext, lib.HashSize)

    	var wg sync.WaitGroup
	    wg.Add(nbGoRoutines)
    	for g := 0; g < nbGoRoutines; g++{

    		start := g*coefficientsPerGoRoutine
    		end := (g+1)*coefficientsPerGoRoutine

    		if g == nbGoRoutines-1{
    			end = lib.HashSize-1
    		}

            go func(worker, startHash, endHash int){

            	startGenome := i*int(c.params.N())
            	endGenome := (i+1)*int(c.params.N())
            	if endGenome > nbGenomes-1{
            		endGenome = nbGenomes-1
            	}

            	tmp := encryptor.Encrypt(worker, startGenome, endGenome, hashTransposed[startHash:endHash])

            	for j := startHash; j < endHash; j++{
            		ciphertexts[i][j] = tmp[j-startHash]
            	}

                wg.Done()
            }(g, start, end)
    	}
    	wg.Wait()


    	// Marshales the ciphertexts

    	
    }
    


}