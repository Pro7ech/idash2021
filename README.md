# EPFL solution for iDash 2021 Task2

## How to build the Docker

```
docker build -t epfl_solution .
docker run -it epfl_solution
```

## Processing Samples

Samples are pre-processed by first applying a FCGR mapping, followed by a 2D DCTII [Lichtblau2019].
The top left h x h matrix of the DCTII (lowest frequencies) is extracted and set at the hash of the genome.

## Training

`$ go run model/main.go` will process the samples of `data/Challenge.fa` and output the processed samples in `model/X.binary` `model/Y.binary` (X being the processed samples and Y the labels).
The Python script `model/training.py` will use `model/X.binary` `model/Y.binary` that can then be used to train the model.
The script will output the weights both in `.npy` and `.binary` as well as a `.png` image of the weights/features with gradient color coding.

## Testing
`$ make debug NBGENOMES=2000` will compile and run `DebugTest.go` which will process, encrypt, predict, decrypt the first 2000 samples located in `data/Challenge.fa`.

## Run iDash21
- `$ make key` : generates the secret-key and stores it in `key/`.
- `$ make pro NBGENOMES=2000` : processes the first 2000 samples located iin `data/Challenge.fa`. Returns the result in `temps/`.
- `$ make enc` : Encrypts the processed samples. Returns the encrypted processed samples in `temp/`.
- `$ make pred` : unmarshals the encrypted samples in  `temp/`, evaluates the homomorphic prediction and marshals back the result in `temp/`.
- `$ make dec` : unmarshals the encrypted prediction in `temp/`, decrypts and outputs the result in `results/prediction.csv`.

## Parameters
Processing and crypto parameters are located in `lib/params`.

## Others
- `$ make clean` : clean all files in `keys/`, `temps/`,`results/` and all compiled binary files. Does not clean files in `model/`.

## Security
The HE evaluation security is based on the R-LWE hardness. The used parameters are log(N)=10, log(Q)=29. Both the secret and the Gaussian error are sampled from a truncated discrete Gaussian distribution with standard deviation 3.19 and bound 19. The security is estimated to 128-bit according to https://homomorphicencryption.org/.


## References:
[Lichtblau2019] : Lichtblau Daniel. “Alignment-free genomic sequence comparison using FCGRand signal processing”, 2019.
