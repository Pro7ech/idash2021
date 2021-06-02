package lib


//Strain name map
var StrainsMap = map[int]string{
    0:"B.1.427",
    1:"B.1.1.7",
    2:"P.1",
    3:"B.1.526",
    }

// General parameters
var NbSamplesPerStrain = 2000
var NbStrains = 4
var NbSamples = NbStrains*NbSamplesPerStrain

// Client pre-processing parameters
var HashSqrtSize = 16
var HashSize = HashSqrtSize*HashSqrtSize
var Window = 8

// Parallelization parameters
var NbGoRoutines = 4

// Crypto parameters
var LogN uint64 = 10
var Q = []uint64{0x20002801}
var HashScale float64 = 1<<16
var ModelScale float64 = 7
var Sigma = 3.2
var SigmaBound uint64 = 19