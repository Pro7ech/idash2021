package preprocessing

import (
	"math"
)

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
	nbGo       int
	window     int
	hsize      int
	normalizer float64
	dct        *ParallelDCTII
	cgrmatrix  []CRGMatrix
	cgrhash    []DCTHash
}

func NewDCTHasher(nbGo, window, hashsqrtsize int, normalizer float64) *DCTHasher {

	dct := NewParallelDCTII(nbGo, 1<<window)

	pool := make([]CRGMatrix, nbGo)
	for i := range pool {
		pool[i] = NewCRGMatrix(window)
	}

	hash := make([]DCTHash, nbGo)
	for i := range hash {
		hash[i] = NewDCTHash(hashsqrtsize)
	}

	return &DCTHasher{
		nbGo:       nbGo,
		window:     window,
		hsize:      hashsqrtsize,
		normalizer: normalizer,
		dct:        dct,
		cgrmatrix:  pool,
		cgrhash:    hash}
}

func (dcth *DCTHasher) Hash(worker int, dna string) {
	dcth.MapCGR(worker, dna)
	dcth.DCTII(worker)
	dcth.Finalize(worker)
}

func (dcth *DCTHasher) MapCGR(worker int, dna string) {

	window := dcth.window
	normalizer := dcth.normalizer
	cgrmatrix := dcth.cgrmatrix[worker]

	//Zero the matrix
	for i := range cgrmatrix {
		tmp := cgrmatrix[i]
		for j := range tmp {
			tmp[j] = 0.0
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

	// Non linear normalization
	for i := range cgrmatrix {
		tmp := cgrmatrix[i]
		for j := range tmp {
			tmp[j] = math.Pow(tmp[j]/max, normalizer)
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
	hsize := dcth.hsize

	dcth.dct.Transform2DToHash(worker, hsize, cgrmatrix)
}

func (dcth *DCTHasher) Finalize(worker int) {

	cgrmatrix := dcth.cgrmatrix[worker]
	hash := dcth.cgrhash[worker]
	hsize := dcth.hsize

	for i := 0; i < hsize; i++ {
		tmp := cgrmatrix[i]
		idx := i * hsize
		for j := 0; j < hsize; j++ {
			hash[idx+j] = tmp[j]
		}
	}
}

type DCTHasherV2 struct {
	nbGo       int
	window     int
	hsize      int
	normalizer float64
	dct        *ParallelDCTII
	mA         [][]float64
	mC         [][]float64
	mG         [][]float64
	mT         [][]float64
	cgrmatrix  []CRGMatrix
	cgrhash    []DCTHash
}

func NewDCTHasherV2(nbGo, window, hashsqrtsize int, normalizer float64) *DCTHasherV2 {

	dct := NewParallelDCTII(nbGo, 2*(1<<(window>>1)))

	pool := make([]CRGMatrix, nbGo)
	mA := make([][]float64, nbGo)
	mC := make([][]float64, nbGo)
	mG := make([][]float64, nbGo)
	mT := make([][]float64, nbGo)
	for i := range pool {
		pool[i] = NewCRGMatrix(1 + (window >> 1))
		mA[i] = make([]float64, 1<<window)
		mC[i] = make([]float64, 1<<window)
		mG[i] = make([]float64, 1<<window)
		mT[i] = make([]float64, 1<<window)
	}

	hash := make([]DCTHash, nbGo)
	for i := range hash {
		hash[i] = NewDCTHash(hashsqrtsize)
	}

	return &DCTHasherV2{
		nbGo:       nbGo,
		window:     window,
		hsize:      hashsqrtsize,
		normalizer: normalizer,
		dct:        dct, cgrmatrix: pool,
		mA:      mA,
		mC:      mC,
		mG:      mG,
		mT:      mT,
		cgrhash: hash}
}

func (dcth *DCTHasherV2) Hash(worker int, dna string) {
	dcth.MapCGR(worker, dna)
	dcth.DCTII(worker)
}

func (dcth *DCTHasherV2) MapCGR(worker int, dna string) {

	window := dcth.window
	cgrmatrix := dcth.cgrmatrix[worker]
	normalizer := dcth.normalizer
	mA := dcth.mA[worker]
	mC := dcth.mA[worker]
	mG := dcth.mA[worker]
	mT := dcth.mA[worker]

	for i := range mA {
		mA[i] = 0.0
		mC[i] = 0.0
		mG[i] = 0.0
		mT[i] = 0.0
	}

	xA, xC, xG, xT := MapSubStringTo4x1D(dna[0 : 0+window])

	mA[xA]++
	mC[xC]++
	mG[xG]++
	mT[xT]++

	mask := 1 << (window - 1)

	for j := window; j < len(dna); j++ {

		xA >>= 1
		xC >>= 1
		xG >>= 1
		xT >>= 1

		char := dna[j]

		if char == 'A' {
			xA |= mask
		} else if char == 'C' {
			xC |= mask
		} else if char == 'G' {
			xG |= mask
		} else if char == 'T' {
			xT |= mask
		} else if char == 'R' {
			xA |= mask
			xG |= mask
		} else if char == 'Y' {
			xC |= mask
			xT |= mask
		} else if char == 'K' {
			xG |= mask
			xT |= mask
		} else if char == 'M' {
			xA |= mask
			xC |= mask
		} else if char == 'S' {
			xC |= mask
			xG |= mask
		} else if char == 'W' {
			xA |= mask
			xT |= mask
		} else if char == 'N' {
			xA |= mask
			xC |= mask
			xG |= mask
			xT |= mask
		}

		mA[xA]++
		mC[xC]++
		mG[xG]++
		mT[xT]++
	}

	maxA := maxSlice(mA)
	maxC := maxSlice(mC)
	maxG := maxSlice(mG)
	maxT := maxSlice(mT)

	for i := range mA {
		mA[i] = math.Pow(mA[i]/maxA, normalizer)
	}

	for i := range mA {
		mC[i] = math.Pow(mC[i]/maxC, normalizer)
	}

	for i := range mA {
		mG[i] = math.Pow(mG[i]/maxG, normalizer)
	}

	for i := range mA {
		mT[i] = math.Pow(mT[i]/maxT, normalizer)
	}

	halfSize := (1 << (window >> 1))

	for i := 0; i < (1 << (window >> 1)); i++ {
		for j := 0; j < (1 << (window >> 1)); j++ {
			cgrmatrix[i][j] = mA[i*halfSize+j]
			cgrmatrix[i][j+halfSize] = mC[i*halfSize+j]
			cgrmatrix[i+halfSize][j] = mG[i*halfSize+j]
			cgrmatrix[i+halfSize][j+halfSize] = mT[i*halfSize+j]
		}
	}
}

func (dcth *DCTHasherV2) DCTII(worker int) {

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

func MapSubStringTo4x1D(substring string) (xA, xC, xG, xT int) {
	i := 1
	for _, char := range substring {
		if char == 'A' {
			xA |= i
		} else if char == 'C' {
			xC |= i
		} else if char == 'G' {
			xG |= i
		} else if char == 'T' {
			xT |= i
		} else if char == 'R' {
			xA |= i
			xG |= i
		} else if char == 'Y' {
			xC |= i
			xT |= i
		} else if char == 'K' {
			xG |= i
			xT |= i
		} else if char == 'M' {
			xA |= i
			xC |= i
		} else if char == 'S' {
			xC |= i
			xG |= i
		} else if char == 'W' {
			xA |= i
			xT |= i
		} else {
			xA |= i
			xC |= i
			xG |= i
			xT |= i
		}
		i <<= 1
	}
	return
}

func (dcth *DCTHasherV2) GetCGR(worker int) [][]float64 {
	return dcth.cgrmatrix[worker]
}

func (dcth *DCTHasherV2) GetHash(worker int) []float64 {
	return dcth.cgrhash[worker]
}

func (dcth *DCTHasher) GetCGR(worker int) [][]float64 {
	return dcth.cgrmatrix[worker]
}

func (dcth *DCTHasher) GetHash(worker int) []float64 {
	return dcth.cgrhash[worker]
}

func maxSlice(slice []float64) (max float64) {
	max = 0.0
	for _, v := range slice {
		if max < v {
			max = v
		}
	}
	return
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
