package lib

import (
	"strconv"
)

var GenomeDataPath = "data/Challenge.fa"

// KeysPath is a variable pointing to the Keys folder
var KeysPath = "keys/SecretKey.binary"

// EncDataPath is a variable pointing to the ciphertext(s) folder
var EncDataPath = "temps/"

// PredDataPath is a variable pointing to model(s) folder
var PredDataPath = "prediction_data/"

// ClientParamsPath is a variable pointing to the client's scheme parameters (marshaled)
var ClientParamsPath = EncDataPath + "enc_client_params.binary"

// ClientEncDataPath is a variable pointing to the client encrypted data (marshaled)
var ClientEncDataPath = EncDataPath + "enc_client_data.binary"

// ClientResDataPath is a variable pointing to the client decrypted prediction data (marshaled)
var ClientResDataPath = "results/ypred.binary"

// ServerMappingTablePath is a variable pointing to the model(s)
var ServerMappingTablePath = PredDataPath + "mapping_table.txt"

// ServerEncParameters is a variable pointing to the processed ciphertexts of the server (marshaled)
var ServerEncParameters = EncDataPath + "enc_pred_parameters.binary"

func EncryptedBatchIndexPath(index int) (string){
	return EncDataPath + "enc_client_batch_" + strconv.Itoa(index) + ".binary"
}
