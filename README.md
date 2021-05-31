# iDash 2021 Task2

## Parameters
- NbGoRoutines : number of parallel workers
- Window : window size of Chaos Game Representation mapping
- HashSqrtSize : dimension of the hash matrix

## Processing Samples

Samples are pre-processed by first applying a CGR mapping (_Genomic Signature: Characterization and Classification of Species Assessed
by Chaos Game Representation of Sequences_), followed by a 2D DCTII (_A perceptual hash function to store and retrieve large scale DNA sequences_).
The top left HashSqrtSize x HashSqrtSize matrix of the DCTII (highest frequencies) is extracted and set at the hash of the genome.

- Create a new `NewDCTHasher(nbgoroutines, window)`
- For each strain call `.Hash(worker_idx, dna_string)`
- Get the hash with `.GetHash(worker_idx)`


## Training

Linear regression of size (HashSqrtSize x HashSqrtSize) -> 4