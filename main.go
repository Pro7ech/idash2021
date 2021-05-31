package main

import (
    "bufio"
    "math"
    "fmt"
    "log"
    "os"
    "encoding/binary"
    "github.com/ldsec/idash21_Task2/preprocessing"
)

//Strain name map
var strains_map = map[int]string{
    0:"B.1.427",
    1:"B.1.1.7",
    2:"P.1",
    3:"B.1.526",
    }



func main(){

    nbSamplesStrain := preprocessing.NbSamplesPerStrain
    nbSamples := preprocessing.NbSamples
    hashsqrtsize := preprocessing.HashSqrtSize
    window := preprocessing.Window
    nbGo := 1

    var err error 
	file, err := os.Open("data/Challenge.fa")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    buff := make([]byte, hashsqrtsize*hashsqrtsize*8)

    // Creates the files containing the processed samples
    var fw *os.File
    if fw, err = os.Create("data/X_CGR_DCT"); err != nil {
        panic(err)
    }

    hasher := preprocessing.NewDCTHasher(nbGo, window, hashsqrtsize)

    i := 0
    for scanner.Scan() {

    	if i&1 == 1{

    		if i%200 == 1{
    			fmt.Printf("%-7s : %4d/%d\n", strains_map[i/(2*nbSamplesStrain)], i>>1, nbSamples)
    		}
    		strain := scanner.Text()
            
            hasher.Hash(0, strain)
            hash := hasher.GetHash(0)
 
            if fw, err = os.OpenFile("data/X_CGR_DCT", os.O_APPEND|os.O_WRONLY, 0644); err != nil{
                panic(err)
            }

            for i := range hash{
                binary.LittleEndian.PutUint64(buff[i<<3:(i+1)<<3], math.Float64bits(hash[i]))
            }
            fw.Write(buff)
            defer file.Close()
    	}  
    	i++
    }
}