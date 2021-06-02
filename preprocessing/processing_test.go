package preprocessing


import(
	"testing"
	"math/rand"
)

func BenchmarkProcessing(b *testing.B){

	hasher := NewDCTHasher(1, 8, 16)

	var runesACGT = []rune("ACGT")
	var runesRYKSMWN = []rune("RYKSMWN")

	//Create a random string of length 28000
	d := make([]rune, 28000)
	for i := range d{

		insert := rand.Float64()
		if insert < 0.01{
			d[i] = runesRYKSMWN[rand.Intn(len(runesRYKSMWN))]
		}else{
			d[i] = runesACGT[rand.Intn(len(runesACGT))]
		}
	}

	strain := string(d)

	b.Run("CGR_256x256", func(b *testing.B) {

		for i := 0; i < b.N; i++{
			hasher.MapCGR(0, strain)
		}

	})

	b.Run("DCTII_256x256_16x16", func(b *testing.B) {

		for i := 0; i < b.N; i++{
			hasher.DCTII(0)
		}

	})
}