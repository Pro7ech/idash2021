package preprocessing

type CRGMatrix [][]float64

func NewCRGMatrix(window int) (cgr CRGMatrix) {
	cgr = make([][]float64, 1<<window)
	for i := range cgr {
		cgr[i] = make([]float64, 1<<window)
	}
	return
}

type DCTHash []float64

func NewDCTHash(HashSqrtSize int) (hash []float64) {
	return make([]float64, HashSqrtSize*HashSqrtSize)
}

type DCTHasher struct {
	nbGo      int
	window    int
	hsize     int
	dct       *ParallelDCTII
	cgrmatrix []CRGMatrix
	cgrhash   []DCTHash
}

func NewDCTHasher(nbGo, window, hashsqrtsize int) *DCTHasher {

	dct := NewParallelDCTII(nbGo, 1<<window)

	pool := make([]CRGMatrix, nbGo)
	for i := range pool {
		pool[i] = NewCRGMatrix(window)
	}

	hash := make([]DCTHash, nbGo)
	for i := range hash {
		hash[i] = NewDCTHash(hashsqrtsize)
	}

	return &DCTHasher{nbGo: nbGo, window: window, hsize: hashsqrtsize, dct: dct, cgrmatrix: pool, cgrhash: hash}
}

func (dcth *DCTHasher) Hash(worker int, dna string) {
	dcth.MapCGR(worker, dna)
	dcth.DCTII(worker)
}

func (dcth *DCTHasher) MapCGR(worker int, dna string) {

	window := dcth.window
	cgrmatrix := dcth.cgrmatrix[worker]

	//Zero the matrix
	for i := range cgrmatrix {
		tmp := cgrmatrix[i]
		for j := range tmp {
			tmp[j] = complex(0, 0)
		}
	}

	var x, y int
	for j := 0; j < len(dna)-window+1; j++ {
		substring := dna[j : j+window]
		x, y = MapSubString2D(substring)
		if x != -1 {
			cgrmatrix[x][y] += 1.0
		} else {
			j += y // skip to the next char after the invalid char
		}
	}

	// Get the maximum value of the matrix
	max := maxDoubleSlice(cgrmatrix)

	// Normalizes by the maximum value
	for i := range cgrmatrix {
		tmp := cgrmatrix[i]
		for j := range tmp {
			tmp[j] /= max
		}
	}
}

func MapSubString2D(substring string) (x, y int) {
	i := 1
	for j, char := range substring {

		if char == 'C' {
			y |= i
		} else if char == 'G' {
			x |= i
		} else if char == 'T' {
			x |= i
			y |= i
		} else if char != 'A' {
			return -1, j // returns that invalid char is present and its index
		}
		i <<= 1
	}
	return x, y
}

func (dcth *DCTHasher) DCTII(worker int) {

	cgrmatrix := dcth.cgrmatrix[worker]
	hash := dcth.cgrhash[worker]
	hsize := dcth.hsize

	dcth.dct.Transform2DToHash(worker, hsize, cgrmatrix)

	for i := 0; i < hsize; i++ {
		tmp := cgrmatrix[i]
		idx := i * hsize
		for j := 0; j < hsize; j++ {
			hash[idx+j] = tmp[j]
		}
	}
}

func (dcth *DCTHasher) GetCGR(worker int) [][]float64 {
	return dcth.cgrmatrix[worker]
}

func (dcth *DCTHasher) GetHash(worker int) []float64 {
	return dcth.cgrhash[worker]
}

func maxDoubleSlice(doubleSlice [][]float64) (max float64) {
	max = 0.0
	for _, slice := range doubleSlice {
		for _, v := range slice {
			if max < v {
				max = v
			}
		}
	}
	return
}
