package lib


import (
	"encoding/binary"
	"errors"
	"math"
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/ring"
)

// GetCiphertextDataLen returns the expected size of a ciphertext in bytes if marshaled
// Set WithMetaData to true if the metadata must be included
func GetCiphertextDataLen(ciphertext *ckks.Ciphertext, WithMetaData bool) (dataLen uint64) {
	if WithMetaData {
		dataLen += 11
	}

	for _, el := range ciphertext.Value() {
		dataLen += el.GetDataLen32(WithMetaData)
	}

	return dataLen
}

// MarshalBinaryCiphertext32 marshals the input ciphertext on the provided slice of bytes
// Returns an error if the target slice of bytes is too small
// Use GetCiphertextDataLen(ciphertext, true) to get the correct size in bytes
func MarshalBinaryCiphertext32(ciphertext *ckks.Ciphertext, data []byte) (err error) {

	data[0] = uint8(ciphertext.Degree() + 1)

	binary.LittleEndian.PutUint64(data[1:9], math.Float64bits(ciphertext.Scale()))

	if ciphertext.IsNTT() {
		data[10] = 1
	}

	var pointer, inc uint64

	pointer = 11

	for _, el := range ciphertext.Value() {

		if inc, err = el.WriteTo32(data[pointer:]); err != nil {
			return err
		}

		pointer += inc
	}

	return nil
}

// UnmarshalBinaryCiphertext32 unmarshals the provided bytes on the input ciphertext
func UnmarshalBinaryCiphertext32(ciphertext *ckks.Ciphertext, data []byte) (err error) {
	if len(data) < 11 {
		return errors.New("too small bytearray")
	}

	ciphertext.SetScale(math.Float64frombits(binary.LittleEndian.Uint64(data[1:9])))

	if uint8(data[10]) == 1 {
		ciphertext.SetIsNTT(true)
	}

	var pointer uint64
	pointer = 11

	for i := range ciphertext.Value() {
		pointer += DecodeCoeffs32(ciphertext.Value()[i].Coeffs, data[pointer:])
	}

	if pointer != uint64(len(data)) {
		return errors.New("remaining unparsed data")
	}

	return nil
}

// GetCiphertextDataLenSeeded returns the expected size in bytes of a ciphertext that was generated
// by a seeded encryption (the uniform polynomial a of [-as + m + e, a] is generated deterministically)
// In this case, the degree 1 element of the ciphertext (the element a) does not need to be stored
func GetCiphertextDataLenSeeded(ciphertext *ckks.Ciphertext, WithMetaData bool) (dataLen uint64) {
	if WithMetaData {
		dataLen += 11
	}

	dataLen += ciphertext.Value()[0].GetDataLen32(WithMetaData)

	return dataLen
}

// MarshalBinaryCiphertextSeeded32 only marshals the degree zero element of the inpu ciphertext on the provided slice of bytes
// Returns an error if the target slice of bytes is too small
// Use GetCiphertextDataLenSeeded(ciphertext, true) to get the correct size in bytes
func MarshalBinaryCiphertextSeeded32(ciphertext *ckks.Ciphertext, data []byte) (err error) {

	// Degree will be read as the mask is not included during the encryption
	// so we add one more, such that during the unmarshaling the correct ciphertext
	// is created
	data[0] = uint8(ciphertext.Degree() + 1 + 1)

	binary.LittleEndian.PutUint64(data[1:9], math.Float64bits(ciphertext.Scale()))

	if ciphertext.IsNTT() {
		data[10] = 1
	}

	var pointer uint64

	pointer = 11

	if _, err = ciphertext.Value()[0].WriteTo32(data[pointer:]); err != nil {
		return err
	}

	return nil
}

// UnmarshalBinaryCiphertextSeeded32 unmarshals the provided bytes on the input ciphertext
// Only the degree zero element is recovered, the degree 1 element needs to be reconstructed
// from the seed
func UnmarshalBinaryCiphertextSeeded32(ciphertext *ckks.Ciphertext, data []byte) (err error) {
	if len(data) < 11 {
		return errors.New("too small bytearray")
	}

	ciphertext.Element = new(ckks.Element)

	ciphertext.SetValue(make([]*ring.Poly, uint8(data[0])))

	ciphertext.SetScale(math.Float64frombits(binary.LittleEndian.Uint64(data[1:9])))

	if uint8(data[10]) == 1 {
		ciphertext.SetIsNTT(true)
	}

	var pointer, inc uint64
	pointer = 11

	ciphertext.Value()[0] = new(ring.Poly)

	if inc, err = ciphertext.Value()[0].DecodePolyNew32(data[pointer:]); err != nil {
		return err
	}

	if pointer+inc != uint64(len(data)) {
		return errors.New("remaining unparsed data")
	}

	return nil
}

// DecodeCoeffs32 converts a byte array to a matrix of coefficients.
func DecodeCoeffs32(coeffs [][]uint64, data []byte) (pointer uint64) {

	N := uint64(1 << data[0])
	numberModuli := uint64(data[1])
	pointer = 2

	tmp := N << 2
	for i := uint64(0); i < numberModuli; i++ {
		for j := uint64(0); j < N; j++ {
			coeffs[i][j] = uint64(binary.BigEndian.Uint32(data[pointer+(j<<2) : pointer+((j+1)<<2)]))
		}
		pointer += tmp
	}

	return pointer
}