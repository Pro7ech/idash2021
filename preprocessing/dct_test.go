package preprocessing

import(
	"testing"
	"math"

	"github.com/ardabasaran/go-fourier"
)

func TestDCTII(t *testing.T){

	t.Run("DCT2D_256x256", func(t *testing.T) {

		n := 256

		dct := NewParallelDCTII(1, n)
		
		matrix := make([][]float64, n)
		for i := range matrix{
			matrix[i] = make([]float64, n)
			for j := range matrix{
				matrix[i][j] = float64((j+i)%8)
			}
		}

		m_want, _ := go_fourier.DCT2D(matrix)

		dct.Transform2D(0, matrix)

		epsilon := 1e-10

		for i := range matrix{
			tmp0 := matrix[i]
			tmp1 := m_want[i]
			for j := range matrix[i]{
				if math.Abs(tmp0[j]-tmp1[j]) > epsilon{
					t.Error("Fail")
				}
			}
		}
	})
}

func BenchmarkDCTII(b *testing.B){

	n := 256

	dct := NewParallelDCTII(1, n)
	
	matrix := make([][]float64, n)
	for i := range matrix{
		matrix[i] = make([]float64, n)
		for j := range matrix{
			matrix[i][j] = float64((j+i)%8)
		}
	}

	b.Run("DCT2D_256x256_custom", func(b *testing.B) {

		for i := 0; i < b.N; i++{
			dct.Transform2D(0, matrix)
		}
	})

	b.Run("DCT2D_256x256_native", func(b *testing.B) {

		for i := 0; i < b.N; i++{
			go_fourier.DCT2D(matrix)
		}
	})
} 