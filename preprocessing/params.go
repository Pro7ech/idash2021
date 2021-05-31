package preprocessing


//Strain name map
var StrainsMap = map[int]string{
    0:"B.1.427",
    1:"B.1.1.7",
    2:"P.1",
    3:"B.1.526",
    }

//Fractal position map
var CoordMap = map[string][2]int{
    "A" : [2]int{0, 0},
    "C" : [2]int{0, 1},
    "G" : [2]int{1, 0},
    "T" : [2]int{1, 1}}

// General parameters
var NbSamplesPerStrain = 2000
var NbStrains = 4
var NbSamples = NbStrains*NbSamplesPerStrain
var HashSqrtSize = 16
var Window = 8