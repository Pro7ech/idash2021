package prediction

import (
	"encoding/binary"
	"fmt"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/ring"
	"math"
	"math/big"
	"os"
	"unsafe"
)

type Predictor struct {
	params   *ckks.Parameters
	baseRing *ring.Ring
	model    *Model
	pool     []*ring.Poly
}

type Model struct {
	weights                 [][]float64
	bias                    []float64
	weightsScaledMontgomery [][]uint64
	biasScaled              []*ring.Poly
}

func NewPredictor(schemeParams *ckks.Parameters) *Predictor {
	ringQ, _ := ring.NewRing(schemeParams.N(), schemeParams.Qi())
	pool := make([]*ring.Poly, lib.NbStrains)
	for i := range pool {
		pool[i] = ringQ.NewPoly()
	}
	return &Predictor{params: schemeParams, baseRing: ringQ, pool: pool}
}

func (p *Predictor) PrintModel() {
	fmt.Println("==================== BIAS ===================")
	fmt.Printf("%3d : ", 0)
	for i := range p.model.bias {
		fmt.Printf("%9.4f ", p.model.bias[i])
	}
	fmt.Printf("\n")

	fmt.Println("================== WEIGHTS ==================")
	for i := range p.model.weights[0] {
		fmt.Printf("%3d : ", i)
		for j := range p.model.weights {
			fmt.Printf("%9.4f ", p.model.weights[j][i])
		}
		fmt.Printf("\n")
	}
}

func (p *Predictor) LoadModel(path string) {

	nbStrains := lib.NbStrains
	baseRing := p.baseRing
	bredParams := baseRing.GetBredParams()[0]
	Q := baseRing.Modulus[0]

	p.model = new(Model)

	var err error
	var fr *os.File
	if fr, err = os.Open(path + "weights_layer_0"); err != nil {
		panic(err)
	}

	frInfo, err := fr.Stat()
	buff := make([]byte, frInfo.Size())
	if _, err := fr.Read(buff); err != nil {
		panic(err)
	}

	weights := make([][]float64, lib.NbStrains)
	weightsScaledMontgomery := make([][]uint64, lib.HashSize)
	for i := range weights {
		tmp0 := make([]float64, lib.HashSize)
		tmp1 := make([]uint64, lib.HashSize)
		for j := range tmp0 {
			tmp0[j] = math.Float64frombits(binary.LittleEndian.Uint64(buff[(i+j*nbStrains)<<3 : (i+j*nbStrains+1)<<3]))
			tmp1[j] = ring.MForm(scaleUpExact(tmp0[j], lib.ModelScale, Q), Q, bredParams)

		}
		weights[i] = tmp0
		weightsScaledMontgomery[i] = tmp1
	}

	p.model.weights = weights
	p.model.weightsScaledMontgomery = weightsScaledMontgomery

	if fr, err = os.Open(path + "bias_layer_0"); err != nil {
		panic(err)
	}

	frInfo, err = fr.Stat()
	buff = make([]byte, frInfo.Size())
	if _, err := fr.Read(buff); err != nil {
		panic(err)
	}

	bias := make([]float64, lib.NbStrains)
	biasScaled := make([]*ring.Poly, lib.NbStrains)
	for i := range bias {
		bias[i] = math.Float64frombits(binary.LittleEndian.Uint64(buff[(i)<<3 : (i+1)<<3]))

		tmp := baseRing.NewPoly()
		baseRing.AddScalar(tmp, scaleUpExact(bias[i], lib.HashScale*lib.ModelScale, Q), tmp)
		baseRing.NTT(tmp, tmp)

		biasScaled[i] = tmp
	}

	p.model.bias = bias
	p.model.biasScaled = biasScaled
}

func (p *Predictor) Predict(input []*ckks.Ciphertext, output []*ckks.Ciphertext) {
	for i := range output {
		p.DotProduct(input, i, output[i])
	}
}

// Multiplies a list of ciphertext with the weights of the given label and sums it all on the output ciphertext
func (p *Predictor) DotProduct(input []*ckks.Ciphertext, labelIndex int, output *ckks.Ciphertext) {

	baseRing := p.baseRing

	weights := p.model.weightsScaledMontgomery[labelIndex]
	bias := p.model.biasScaled[labelIndex]
	pool := p.pool[labelIndex]

	for i := range input {

		weight := weights[i]

		p0 := input[i].Value()[0].Coeffs[0]
		p1 := pool.Coeffs[0]

		// Montgomery multiplication without modular reduction
		// sum(ai * 2^64 * bi) = 2^64 * sum(ai * bi)
		for j := uint64(0); j < baseRing.N; j = j + 8 {

			x := (*[8]uint64)(unsafe.Pointer(&p0[j]))
			y := (*[8]uint64)(unsafe.Pointer(&p1[j]))

			y[0] += x[0] * weight
			y[1] += x[1] * weight
			y[2] += x[2] * weight
			y[3] += x[3] * weight
			y[4] += x[4] * weight
			y[5] += x[5] * weight
			y[6] += x[6] * weight
			y[7] += x[7] * weight
		}

		if i%64 == 63 || i == len(input)-1 {
			baseRing.InvMForm(pool, pool)
			baseRing.Add(output.Value()[0], pool, output.Value()[0])
			pool.Zero()
		}
	}

	for i := range input {

		weight := weights[i]

		p0 := input[i].Value()[1].Coeffs[0]
		p1 := pool.Coeffs[0]

		// Montgomery multiplication without modular reduction
		// sum(ai * 2^64 * bi) = 2^64 * sum(ai * bi)
		for j := uint64(0); j < baseRing.N; j = j + 8 {

			x := (*[8]uint64)(unsafe.Pointer(&p0[j]))
			y := (*[8]uint64)(unsafe.Pointer(&p1[j]))

			y[0] += x[0] * weight
			y[1] += x[1] * weight
			y[2] += x[2] * weight
			y[3] += x[3] * weight
			y[4] += x[4] * weight
			y[5] += x[5] * weight
			y[6] += x[6] * weight
			y[7] += x[7] * weight
		}

		if i%64 == 63 || i == len(input)-1 {
			baseRing.InvMForm(pool, pool)
			baseRing.Add(output.Value()[1], pool, output.Value()[1])
			pool.Zero()
		}
	}

	baseRing.Add(output.Value()[0], bias, output.Value()[0])
}

// Returns value * n mod Q
func scaleUpExact(value float64, n float64, q uint64) (res uint64) {

	var isNegative bool
	var xFlo *big.Float
	var xInt *big.Int

	isNegative = false
	if value < 0 {
		isNegative = true
		xFlo = big.NewFloat(-n * value)
	} else {
		xFlo = big.NewFloat(n * value)
	}

	xFlo.Add(xFlo, big.NewFloat(0.5))

	xInt = new(big.Int)
	xFlo.Int(xInt)
	xInt.Mod(xInt, ring.NewUint(q))

	res = xInt.Uint64()

	if isNegative {
		res = q - res
	}

	return
}
