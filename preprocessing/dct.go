package preprocessing

import (
	"math"
	"unsafe"
)

type ParallelDCTII struct {
	n       int
	roots   []complex128
	scaling []complex128
	pool    [][]complex128
}

func NewParallelDCTII(nbGo, n int) (dct *ParallelDCTII) {

	if n&(n-1) != 0 {
		panic("n must be a power of two")
	}

	roots := make([]complex128, n)
	for i := range roots {
		angle := 2 * 3.141592653589793 * float64(i) / float64(n)
		roots[i] = complex(math.Cos(angle), math.Sin(angle))
	}

	scaling := make([]complex128, n)
	for i := range scaling {
		angle := 3.141592653589793 * float64(i) / float64(2*n)
		scaling[i] = complex(math.Cos(angle), math.Sin(angle))

		if i == 0 {
			scaling[i] *= complex(math.Sqrt(1/float64(n)), 0)
		} else {
			scaling[i] *= complex(math.Sqrt(2/float64(n)), 0)
		}
	}

	pool := make([][]complex128, nbGo)
	for i := range pool {
		pool[i] = make([]complex128, n)
	}

	return &ParallelDCTII{n: n, roots: roots, scaling: scaling, pool: pool}
}

func (dct *ParallelDCTII) Transform2D(worker int, matrix [][]float64) {

	// Transpose
	for i := 0; i < len(matrix)-1; i++ {
		for j := i + 1; j < len(matrix); j++ {
			matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j]
		}
	}

	// DCT II
	for i := range matrix {
		dct.Transform1D(worker, matrix[i])
	}

	// Transpose
	for i := 0; i < len(matrix)-1; i++ {
		for j := i + 1; j < len(matrix); j++ {
			matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j]
		}
	}

	// DCT II
	for i := range matrix {
		dct.Transform1D(worker, matrix[i])
	}
}

func (dct *ParallelDCTII) Transform1D(worker int, vec []float64) {

	if len(vec) > dct.n {
		panic("vector too large for DCTII parameters")
	}

	if worker > len(dct.pool)-1 {
		panic("#worker larger than DCTII parameters")
	}

	roots := dct.roots
	pool := dct.pool[worker]
	scaling := dct.scaling

	for i := 0; i < len(vec)>>1; i++ {
		pool[i] = complex(vec[i*2], 0)
		pool[dct.n-1-i] = complex(vec[i*2+1], 0)
	}

	sliceBitReverseInPlaceComplex128(pool, dct.n)

	var halfm, gap int
	for m := 2; m <= dct.n; m <<= 1 {
		halfm = m >> 1
		gap = dct.n / m

		if m == 2 {

			psi0 := roots[0*gap]

			for i := 0; i < dct.n; i += 16 {
				x := (*[16]complex128)(unsafe.Pointer(&pool[i]))
				y := (*[16]complex128)(unsafe.Pointer(&pool[i]))

				x[0], y[1] = butterfly(x[0], y[1], psi0)
				x[2], y[3] = butterfly(x[2], y[3], psi0)
				x[4], y[5] = butterfly(x[4], y[5], psi0)
				x[6], y[7] = butterfly(x[6], y[7], psi0)
				x[8], y[9] = butterfly(x[8], y[9], psi0)
				x[10], y[11] = butterfly(x[10], y[11], psi0)
				x[12], y[13] = butterfly(x[12], y[13], psi0)
				x[14], y[15] = butterfly(x[14], y[15], psi0)
			}

		} else if m == 4 {

			psi0 := roots[0*gap]
			psi1 := roots[1*gap]

			for i := 0; i < dct.n; i += 16 {

				x := (*[16]complex128)(unsafe.Pointer(&pool[i]))
				y := (*[16]complex128)(unsafe.Pointer(&pool[i]))

				x[0], y[2] = butterfly(x[0], y[2], psi0)
				x[1], y[3] = butterfly(x[1], y[3], psi1)
				x[4], y[6] = butterfly(x[4], y[6], psi0)
				x[5], y[7] = butterfly(x[5], y[7], psi1)
				x[8], y[10] = butterfly(x[8], y[10], psi0)
				x[9], y[11] = butterfly(x[9], y[11], psi1)
				x[12], y[14] = butterfly(x[12], y[14], psi0)
				x[13], y[15] = butterfly(x[13], y[15], psi1)
			}

		} else if m == 8 {

			psi0 := roots[0*gap]
			psi1 := roots[1*gap]
			psi2 := roots[2*gap]
			psi3 := roots[3*gap]

			for i := 0; i < dct.n; i += 16 {

				x := (*[16]complex128)(unsafe.Pointer(&pool[i]))
				y := (*[16]complex128)(unsafe.Pointer(&pool[i]))

				x[0], y[4] = butterfly(x[0], y[4], psi0)
				x[1], y[5] = butterfly(x[1], y[5], psi1)
				x[2], y[6] = butterfly(x[2], y[6], psi2)
				x[3], y[7] = butterfly(x[3], y[7], psi3)
				x[8], y[12] = butterfly(x[8], y[12], psi0)
				x[9], y[13] = butterfly(x[9], y[13], psi1)
				x[10], y[14] = butterfly(x[10], y[14], psi2)
				x[11], y[15] = butterfly(x[11], y[15], psi3)
			}

		} else {
			for i := 0; i < dct.n; i += m {
				for j := 0; j < halfm; j += 8 {

					x := (*[8]complex128)(unsafe.Pointer(&pool[i+j]))
					y := (*[8]complex128)(unsafe.Pointer(&pool[i+halfm+j]))

					x[0], y[0] = butterfly(x[0], y[0], roots[(j+0)*gap])
					x[1], y[1] = butterfly(x[1], y[1], roots[(j+1)*gap])
					x[2], y[2] = butterfly(x[2], y[2], roots[(j+2)*gap])
					x[3], y[3] = butterfly(x[3], y[3], roots[(j+3)*gap])
					x[4], y[4] = butterfly(x[4], y[4], roots[(j+4)*gap])
					x[5], y[5] = butterfly(x[5], y[5], roots[(j+5)*gap])
					x[6], y[6] = butterfly(x[6], y[6], roots[(j+6)*gap])
					x[7], y[7] = butterfly(x[7], y[7], roots[(j+7)*gap])
				}
			}
		}
	}

	for i := range vec {
		vec[i] = real(pool[i] * scaling[i])
	}
}

func butterfly(x, y, psi complex128) (u, v complex128) {
	tmp := y * psi
	return x + tmp, x - tmp
}

func sliceBitReverseInPlaceComplex128(slice []complex128, n int) {

	var bit, j int

	for i := 1; i < n; i++ {

		bit = n >> 1

		for j >= bit {
			j -= bit
			bit >>= 1
		}

		j += bit

		if i < j {
			slice[i], slice[j] = slice[j], slice[i]
		}
	}
}
