package main

import (
    "bufio"
    "math"
    "fmt"
    "log"
    "os"
    "strings"
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

//Fractal position map
var xy_map = map[string][2]int{
    "A" : [2]int{0, 0},
    "C" : [2]int{0, 1},
    "G" : [2]int{1, 0},
    "T" : [2]int{1, 1}}

// General parameters
var nb_samples_per_strain = 2000
var nb_strains = 4
var nb_samples = nb_strains*nb_samples_per_strain
var hash_size = 16
var window = 8

func NewCRGMatrix(window int)(img [][]float64){
    img = make([][]float64, 1<<window)
    for i := range img{
        img[i] = make([]float64, 1<<window)
    }
    return
}

func MaxDoubleSlice(doubleSlice [][]float64)(max float64){
    max = 0.0
    for _, slice := range doubleSlice{
        for _, v:= range slice{
            if max < v{
                max = v
            }
        }
    }
    return
}

func main(){

    DCTII := preprocessing.NewDCTII(hash_size*hash_size)

    var err error 
	file, err := os.Open("data/Challenge.fa")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    cord_map := map[string][2]int{}

    hash := make([]float64, hash_size*hash_size)
    buff := make([]byte, hash_size*hash_size*8)

    // Creates the files containing the processed samples
    var fw *os.File
    if fw, err = os.Create("data/X_CGR_DCT"); err != nil {
        panic(err)
    }

    i := 0
    nb := 0
    for scanner.Scan() {

    	if i&1 == 1{

            crgMatrix := NewCRGMatrix(window)

    		if i%200 == 1{
    			fmt.Printf("%-7s : %4d/%d\n", strains_map[i/(2*nb_samples_per_strain)], i>>1, nb_samples)
    		}
    		strain := scanner.Text()

            for j := 0; j < len(strain) - window+1; j++{

                substring := strain[j:j+window]

                if strings.ContainsAny(substring, "RYKSMWN"){
                    nb++
                    continue
                }

                var x, y int
                if _, ok := cord_map[substring]; !ok {

                    for i, char := range substring{
                        pos := xy_map[string(char)]
                        x += pos[0]<<i
                        y += pos[1]<<i
                    }
                    cord_map[substring] = [2]int{x, y}
                }else{

                    pos := cord_map[substring]
                    x, y = pos[0], pos[1] 
                }

                crgMatrix[x][y] += 1.0
            }

            // Get the maximum value of the matrix
            max := MaxDoubleSlice(crgMatrix)

            // Normalizes by the maximum value
            for i := range crgMatrix{
                tmp := crgMatrix[i]
                for j := range tmp{
                    tmp[j] /= max
                }
            }

            DCTII.Transform2D(crgMatrix)

            for i := 0; i < hash_size; i++{
                for j := 0; j < hash_size; j++{
                    hash[i*hash_size+j] = crgMatrix[j][i]
                }
            }

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
    fmt.Println(nb)
}