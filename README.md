# iDash 2021 Task2

## Processing Training Samples

Samples are pre-processed by first applying a FCGR mapping (_Genomic Signature: Characterization and Classification of Species Assessed
by Chaos Game Representation of Sequences_), followed by a 2D DCTII (_A perceptual hash function to store and retrieve large scale DNA sequences_).
The top left HashSqrtSize x HashSqrtSize matrix of the DCTII (highest frequencies) is extracted and set at the hash of the genome.

`$ go run model/main.go` will process the samples the results in `model/`.
The processing parameters can be set in `lib/params.go`.

## Training

Run the python script in `model/training.py`, adapt the training parameters as see fit.
The script will output the weights both in `.npy` and `.binary` as well as a `.png` image
of the weights/features with gradiant color coding.

## Testing
`$ make debug NBGENOMES=2000` will compile and run `DebugTest.go` which will process, encrypt, predict, decrypt the first 2000 samples located in `data/Challenge.fa`.

## Run iDash21
- `$ make key` : generate the secret-key, stores it in `key/`.
- `$ make enc NBGENOMES=2000` : process,  encrypts and marshals the first 2000 samples located in `data/Challenge.fa`.
- `$ make pred` : unmarshal the encrypted samples, homomorphic prediction and marshalling of the result.
- `$ make dec` : unmarshals, decrypts and outputs the result in `results/prediction.csv`.

## Others
- `$ make clean` : clean all files in `keys/`, `temps/`,`results/` and all compiled binary files. Does not clean files in `model/`.