package preprocessing

import(
	"strings"
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
	cgrmap map[string][2]int
	dct *ParallelDCTII
	cgrmatrix []CRGMatrix
	cgrhash []DCTHash
}

func NewDCTHasher(nbGo, window, hashsqrtsize int) (*DCTHasher){

	dct := NewParallelDCTII(nbGo, 4*window*window)

	pool := make([]CRGMatrix, nbGo)
	for i := range pool{
		pool[i] = NewCRGMatrix(window)
	}

	hash := make([]DCTHash, nbGo)
	for i := range hash{
		hash[i] = NewDCTHash(hashsqrtsize)
	}

	return &DCTHasher{nbGo:nbGo, window:window, hsize:hashsqrtsize, cgrmap:make(map[string][2]int), dct:dct, cgrmatrix:pool, cgrhash:hash}
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

        if strings.ContainsAny(substring, "RYKSMWN"){
            continue
        }

        var x, y int
        if _, ok := cgrmap[substring]; !ok {

            for i, char := range substring{
                pos := CoordMap[string(char)]
                x += pos[0]<<i
                y += pos[1]<<i
            }
            cgrmap[substring] = [2]int{x, y}
        }else{

            pos := cgrmap[substring]
            x, y = pos[0], pos[1] 
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