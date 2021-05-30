package preprocessing

import(
	"math"
)


type DCTII struct{
	n int
	roots []complex128
	scaling []complex128
	pool []complex128
}


func NewDCTII(n int) (dct *DCTII){
	roots := make([]complex128, n)
	for i := range roots{
		angle := 2*3.141592653589793 * float64(i) / float64(n)
		roots[i] = complex(math.Cos(angle), math.Sin(angle))
	}

	scaling := make([]complex128, n)
	for i := range scaling{
		angle := 3.141592653589793 * float64(i) / float64(2*n) 
		scaling[i] = complex(math.Cos(angle), math.Sin(angle))

		if i == 0{
			scaling[i] *= complex(math.Sqrt(1/float64(n)), 0)
		}else{
			scaling[i] *= complex(math.Sqrt(2/float64(n)),0)
		}
	}

	return &DCTII{n:n, roots:roots, scaling:scaling, pool:make([]complex128, n)}
}


func (dct *DCTII) Transform2D(matrix [][]float64){

	// Transpose
	for i := 0; i < len(matrix)-1; i++{
		for j := i+1; j < len(matrix); j++{
			matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j] 
		}
	}

	// DCT II
	for i := range matrix{
		dct.Transform1D(matrix[i])
	}

	// Transpose
	for i := 0; i < len(matrix)-1; i++{
		for j := i+1; j < len(matrix); j++{
			matrix[i][j], matrix[j][i] = matrix[j][i], matrix[i][j] 
		}
	}

	// DCT II
	for i := range matrix{
		dct.Transform1D(matrix[i])
	}
}


func (dct *DCTII) Transform1D(vec []float64){
	if len(vec) != dct.n{
		panic("vector length not equal to DCTII parameters")
	}

	roots := dct.roots
	pool := dct.pool
	scaling := dct.scaling

	for i := 0; i < dct.n>>1; i++{
		pool[i] = complex(vec[i*2], 0)
		pool[dct.n-1-i] = complex(vec[i*2+1], 0)
	}
	
	sliceBitReverseInPlaceComplex128(pool, dct.n)

	var halfm, gap int
	for m := 2; m <= dct.n; m<<=1{
		halfm = m>>1
		gap = dct.n/m
		for i := 0; i < dct.n; i+=m{
			for k, j := 0, i; j < i+halfm; k,j = k+gap, j+1{
				tmp := pool[j+halfm] * roots[k]
				pool[j+halfm] = pool[j]-tmp
				pool[j] += tmp
			}
		}
	}

	for i := range vec{
		vec[i] = real(pool[i] * scaling[i])
	}
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