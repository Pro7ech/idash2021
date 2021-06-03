package preprocessing

import (
	"math"
	"testing"

	"github.com/ardabasaran/go-fourier"
)

func Test1DDCTII(t *testing.T) {
	t.Run("DCT2D_256x256", func(t *testing.T) {

		n := 64

		dct := NewParallelDCTII(1, n)

		vec0 := make([]float64, n)
		for i := range vec0 {
			vec0[i] = float64((i) % 8)
		}

		vec1, _ := go_fourier.DCT1D(vec0)

		dct.Transform1D(0, vec0)

		epsilon := 1e-10

		for i := range vec0 {
			if math.Abs(vec0[i]-vec1[i]) > epsilon {
				t.Error("Fail")
			}
		}
	})
}

func Test2DDCTII(t *testing.T) {

	t.Run("DCT2D_256x256", func(t *testing.T) {

		n := 16

		dct := NewParallelDCTII(1, n)

		matrix := make([][]float64, n)
		for i := range matrix {
			matrix[i] = make([]float64, n)
			for j := range matrix {
				matrix[i][j] = float64((j + i) % 8)
			}
		}

		m_want, _ := go_fourier.DCT2D(matrix)

		dct.Transform2D(0, matrix)

		epsilon := 1e-10

		for i := range matrix {
			tmp0 := matrix[i]
			tmp1 := m_want[i]
			for j := range matrix[i] {
				if math.Abs(tmp0[j]-tmp1[j]) > epsilon {
					t.Error("Fail")
				}
			}
		}
	})
}

func Benchmark2DDCTII(b *testing.B) {

	n := 256

	dct := NewParallelDCTII(1, n)

	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
		for j := range matrix {
			matrix[i][j] = float64((j + i) % 8)
		}
	}

	b.Run("DCT2D_256x256_custom", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			dct.Transform2D(0, matrix)
		}
	})

	b.Run("DCT2D_256x256_native", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			go_fourier.DCT2D(matrix)
		}
	})
}
