package acc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"

	uuid "github.com/satori/go.uuid"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
)

// KeyId is the key derivation method.
const KeyId = "scrypt"

// DataId is the encryption method used.
const DataId = "aes-256-cbc"

// KeyStore holds values to store account keys.
type KeyStore struct {
	Id  string `json:"id"`
	Ver uint64 `json:"ver"`

	Key  KeyInfo `json:"key"`
	Data KeyData `json:"data"`
	Type KeyType `json:"type"`
}

// KeyInfo holds the key derivation values.
// S   32 (salt)
// N 2048 (cpu cost)
// R    8 (blocksize)
// P    1 (parallelize)
// l   32 (dklen)
type KeyInfo struct {
	T string `json:"t"`
	S string `json:"s"`
	N uint64 `json:"n"`
	P uint64 `json:"p"`
	R uint64 `json:"r"`
	L uint64 `json:"l"`
}

// KeyData holds the encrypted data and the initial vector.
type KeyData struct {
	T string `json:"t"`
	S string `json:"s"`
	D string `json:"d"`
}

// KeyTypes holds the type of key derivation and cipher methodes.
type KeyType struct {
	K string `json:"k"`
	D string `json:"d"`
}

func SaveAccountToFile(acc crypto.Account, pass, path string) error {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("failed to create random iv: %s", err)
	}
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to create random salt: %s", err)
	}

	key, err := scrypt.Key([]byte(pass), salt, 2048, 1, 8, 32)
	if err != nil {
		return fmt.Errorf("failed to create scrypt key: %s", err)
	}
	data, err := mnemonic.FromPrivateKey(acc.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create random salt: %s", err)
	}

	encoded, err := aesEncrypt(data, key, iv, aes.BlockSize)
	if err != nil {
		return fmt.Errorf("failed to create encrypt: %s", err)
	}

	out, err := json.MarshalIndent(KeyStore{
		Id:  uuid.NewV4().String(),
		Ver: 1,
		Key: KeyInfo{
			S: hex.EncodeToString(salt),
			N: 2048,
			R: 8,
			P: 1,
			L: 32,
		},
		Data: KeyData{
			S: hex.EncodeToString(iv),
			D: encoded,
		},
		Type: KeyType{
			K: KeyId,
			D: DataId,
		},
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %s", err)
	}

	fl, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fl.Close()
	fl.Write(out)

	return nil
}

func LoadAccountFromFile(pass, path string) (crypto.Account, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return crypto.Account{}, err
	}

	store := KeyStore{}
	if err = json.Unmarshal(in, &store); err != nil {
		return crypto.Account{}, err
	}
	if store.Type.K != KeyId {
		return crypto.Account{}, fmt.Errorf("unsupported key type: %s", store.Type.K)
	}
	if store.Type.D != DataId {
		return crypto.Account{}, fmt.Errorf("unsupported data type: %s", store.Type.D)
	}

	iv, err := hex.DecodeString(store.Data.S)
	if err != nil {
		return crypto.Account{}, err
	}
	salt, err := hex.DecodeString(store.Key.S)
	if err != nil {
		return crypto.Account{}, err
	}

	key, err := scrypt.Key(
		[]byte(pass), salt, int(store.Key.N),
		int(store.Key.P), int(store.Key.R), int(store.Key.L),
	)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("failed to create scrypt key: %s", err)
	}

	data, err := aesDecrypt(store.Data.D, key, iv)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("failed to create encrypt: %s", err)
	}

	priv, err := mnemonic.ToPrivateKey(data)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("failed to recover key: %s", err)
	}
	acc, err := crypto.AccountFromPrivateKey(priv)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("failed to recover account: %s", err)
	}
	return acc, nil
}

func aesEncrypt(data string, key, iv []byte, blockSize int) (string, error) {
	plainData := addPadding([]byte(data), blockSize, len(data))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("faild to set key: %s", err)
	}

	cipherData := make([]byte, len(plainData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherData, plainData)

	return hex.EncodeToString(cipherData), nil
}

func aesDecrypt(data string, key, iv []byte) (string, error) {
	cipherData, err := hex.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("faild to decode data: %s", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("faild to set key: %s", err)
	}

	plainData := make([]byte, len(cipherData))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainData, cipherData)

	return string(clrPadding(plainData)), nil
}

// clrPadding removes padding from data following PKCS#7
func clrPadding(data []byte) []byte {
	length := int(data[len(data)-1])
	return data[:(len(data) - length)]
}

// addPadding padds the data following the PKCS#7 standard
func addPadding(data []byte, size int, after int) []byte {
	padding := (size - len(data)%size)
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}
