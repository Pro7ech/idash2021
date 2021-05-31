package preprocessing

import(
	"strings"
	"sync"
)

type CRGMatrix [][]float64

func NewCRGMatrix(window int)(cgr CRGMatrix){
    cgr = make([][]float64, 1<<window)
    for i := range cgr{
        cgr[i] = make([]float64, 1<<window)
    }
    return
}

type DCTHash []float64

func NewDCTHash(HashSqrtSize int)(hash []float64){
	return make([]float64, HashSqrtSize*HashSqrtSize)
}

type DCTHasher struct{
	nbGo int
	window int
	hsize int 
	cgrmap *sync.Map
	dct *ParallelDCTII
	cgrmatrix []CRGMatrix
	cgrhash []DCTHash
}

func NewDCTHasher(nbGo, window, hashsqrtsize int) (*DCTHasher){

	dct := NewParallelDCTII(nbGo, 1<<window)

	pool := make([]CRGMatrix, nbGo)
	for i := range pool{
		pool[i] = NewCRGMatrix(window)
	}

	hash := make([]DCTHash, nbGo)
	for i := range hash{
		hash[i] = NewDCTHash(hashsqrtsize)
	}

	return &DCTHasher{nbGo:nbGo, window:window, hsize:hashsqrtsize, cgrmap:new(sync.Map), dct:dct, cgrmatrix:pool, cgrhash:hash}
}

func (dcth *DCTHasher) Hash(worker int, dna string){
	dcth.MapCGR(worker, dna)
	dcth.DCTII(worker)
}

func (dcth *DCTHasher) MapCGR(worker int, dna string){

	window := dcth.window
	cgrmap := dcth.cgrmap
	cgrmatrix := dcth.cgrmatrix[worker]

	//Zero the matrix
	for i := range cgrmatrix{
		tmp := cgrmatrix[i]
		for j := range tmp{
			tmp[j] = complex(0, 0)
		}
	}

	for j := 0; j < len(dna) - window + 1; j++{

        substring := dna[j:j+window]

        // Checks if an invalid character is present
        // If yes, then jump by the position of the invalid character in the 
        // current substring and start at the begining of the loop
    	if idx := strings.IndexAny(substring, "RYKSMWN"); idx != -1{
    		j+=idx
            continue
        }

        var x, y int
        pos, ok := cgrmap.Load(substring)
        if !ok {
            for i, char := range substring{
                xy := CoordMap[string(char)]
                x += xy[0]<<i
                y += xy[1]<<i
            }
            cgrmap.Store(substring, [2]int{x, y})
        }else{
        	x,y = pos.([2]int)[0], pos.([2]int)[1]
        }

        cgrmatrix[x][y] += 1.0
    }

    // Get the maximum value of the matrix
    max := maxDoubleSlice(cgrmatrix)

    // Normalizes by the maximum value
    for i := range cgrmatrix{
        tmp := cgrmatrix[i]
        for j := range tmp{
            tmp[j] /= max
        }
    }
}

func (dcth *DCTHasher) DCTII(worker int){

	cgrmatrix := dcth.cgrmatrix[worker]
	hash := dcth.cgrhash[worker]
	hsize := dcth.hsize

	dcth.dct.Transform2D(worker, cgrmatrix)

    for i := 0; i < hsize; i++{
    	tmp := cgrmatrix[i]
    	idx := i*hsize
        for j := 0; j < hsize; j++{
            hash[idx+j] = tmp[j]
        }
    }
}

func (dcth *DCTHasher) GetCGR(worker int) ([][]float64){
	return dcth.cgrmatrix[worker]
}

func (dcth *DCTHasher) GetHash(worker int) ([]float64){
	return dcth.cgrhash[worker]
}

func maxDoubleSlice(doubleSlice [][]float64)(max float64){
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