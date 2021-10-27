package lib

//Strain name map
var StrainsMap = map[string]int{
	"B.1.427": 0,
	"B.1.1.7": 1,
	"P.1":     2,
	"B.1.526": 3,
}

// Parameters for the training
var NbSamplesPerStrain = 500
var NbStrains = 4
var NbSamples = NbStrains * NbSamplesPerStrain

// Client pre-processing parameters
var HashSqrtSize = 16                      // Dimension of the hash matrix
var HashSize = HashSqrtSize * HashSqrtSize // Number of coefficients in the hash matrix
var Window = 6                             // Fractal Chaos Game Representation window

var Normalizer = 1.0 / 5.0 // Applies x^(normalizer) to the coefficients of the Fractal Chaos Game Representation

// Parallelization parameters
var NbGoRoutines = 4

// Crypto parameters
var LogN uint64 = 10
var Q = []uint64{0x20002801}
var HashScale float64 = 1 << 15
var ModelScale float64 = 7
var Sigma = 3.2
var SigmaBound uint64 = 19
