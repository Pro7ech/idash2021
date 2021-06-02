package client

import(
	"log"
	"crypto/rand"
	"github.com/ldsec/idash21_Task2/lib"
	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/utils"
)


// Encryptor is a struct storing the necessary objects and data to encode the patient data on a plaintext and and encrypt it.
type Encryptor struct {
	params    *ckks.Parameters
	sk        *ring.Poly
	baseRing  *ring.Ring
	thread    []*encryptorThread
}

type encryptorThread struct{
	encoder  ckks.Encoder
	tmpPt    *ckks.Plaintext
	crpGen   *ring.UniformSampler
	gauGen   *ring.GaussianSampler
	seed 	 []byte
	seeded   bool
	pool     *ring.Poly
}

func (enc *Encryptor) Seed(){
	for i := range enc.thread{
		seed := make([]byte, 64)
		if _, err := rand.Read(seed); err != nil {
			log.Fatal(err)
		}

		prngUniform, err := utils.NewKeyedPRNG(seed)
		if err != nil {
			log.Fatal(err)
		}

		enc.thread[i].seed = seed
		enc.thread[i].crpGen = ring.NewUniformSampler(prngUniform, enc.baseRing)
		enc.thread[i].seeded = true
	}
}

func (enc *Encryptor) GetSeeds() (seeds [][]byte){
	seeds = make([][]byte, len(enc.thread))
	for i := range seeds{
		seeds[i] = make([]byte, len(enc.thread[i].seed))
		copy(seeds[i], enc.thread[i].seed)
	}
	return
}

func (enc *Encryptor) newEncryptorThread() (*encryptorThread){
	encoder := ckks.NewEncoder(enc.params)
	tmpPt := ckks.NewPlaintext(enc.params, 0, enc.params.Scale())

	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}

	prngGaussian, err := utils.NewKeyedPRNG(bytes)
	if err != nil {
		log.Fatal(err)
	}

	gauGen := ring.NewGaussianSampler(prngGaussian, enc.baseRing, lib.Sigma, lib.SigmaBound)

	pool := enc.baseRing.NewPoly()

	return &encryptorThread{encoder:encoder, tmpPt:tmpPt, gauGen:gauGen, pool:pool}

}

// NewEncryptor creates a new Encryptor which is thread safe.
func (c *Client) NewEncryptor(nbGoRoutines int) (enc *Encryptor) {
	var err error 

	enc = new(Encryptor)

	enc.params = c.params

	enc.sk = c.sk.Get().CopyNew()

	if enc.baseRing, err = ring.NewRing(c.params.N(), c.params.Qi()); err != nil {
		log.Fatal(err)
	}

	enc.thread = make([]*encryptorThread, nbGoRoutines)
	for i := range enc.thread{
		enc.thread[i] = enc.newEncryptorThread()
	}

	return
}

// Encrypt encodes and encrypts list of slices of float64
func (enc *Encryptor) Encrypt(worker, start, end int, values [][]float64) (ciphertexts []*ckks.Ciphertext) {


	if !enc.thread[worker].seeded{
		panic("encryptor must be seeded to be able to encrypt")
	}

	baseRing := enc.baseRing
	encoder := enc.thread[worker].encoder
	tmpPt   := enc.thread[worker].tmpPt
	crpGen  := enc.thread[worker].crpGen
	gauGen  := enc.thread[worker].gauGen
	pool    := enc.thread[worker].pool


	// Ciphertexts pool
	ciphertexts = make([]*ckks.Ciphertext, len(values))

	var tmpCt *ckks.Ciphertext

	// Generate patient encrypted data for multiplying Ri
	for i := range values {

		// Encodes the vector on the plaintext m
		tmpPt.Value()[0].Zero()
		encoder.EncodeCoeffs(values[i][start:end], tmpPt)

		// Creates a ciphertext of degree 0 (only the first element needs to be stored as the second element is generated from a seed)
		tmpCt = &ckks.Ciphertext{Element: &ckks.Element{}}
		tmpCt.SetScale(enc.params.Scale())
		tmpCt.SetValue(make([]*ring.Poly, 1))
		tmpCt.Value()[0] = baseRing.NewPoly()

		// Encrypts the plaintext on the ciphertext :
		// ct1 = a (generated from the PRNG, only the seed of the PRNG is stored)
		// ct0 = -a * sk + e + m

		// samples NTT(a)
		crpGen.Read(pool)

		// comptues NTT(-a*sk)
		baseRing.MulCoeffsMontgomeryAndSub(pool, enc.sk, tmpCt.Value()[0])

		// computes e + m
		// V1 (uses threadsafe PRNG for gaussian sampling):
		gauGen.ReadAndAdd(tmpPt.Value()[0])

		// NTT(e+m)
		baseRing.NTT(tmpPt.Value()[0], tmpPt.Value()[0])

		// computes NTT(-a *sk) + NTT(e + m)
		baseRing.Add(tmpCt.Value()[0], tmpPt.Value()[0], tmpCt.Value()[0])

		ciphertexts[i] = tmpCt
	}

	return
}
