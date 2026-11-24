package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	
	"golang.org/x/crypto/ripemd160"

	"bytes"
	"encoding/gob"
	"crypto/x509"
)

const version = byte(0x00)
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	
	return *private, pubKey
}

func GetPubKeyHashFromAddress(address []byte) []byte {
	decoded := Base58Decode(address)
	return decoded[1:len(decoded) - addressChecksumLen]
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address 
}

func ValidateAddress(address string) bool {
	decoded := Base58Decode([]byte(address))
	actualChecksum := decoded[len(decoded) - addressChecksumLen :]
	payload := decoded[: len(decoded) - addressChecksumLen]

	targetChecksum := checksum(payload)
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	RIPEMD160Hasher.Write(publicSHA256[:])
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func (w Wallet) GobEncode() ([]byte, error) {
	keyBytes, err := x509.MarshalECPrivateKey(&w.PrivateKey) // or w.PrivateKey if pointer
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(keyBytes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	var keyBytes []byte
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&keyBytes); err != nil {
		return err
	}
	priv, err := x509.ParseECPrivateKey(keyBytes)
	if err != nil {
		return err
	}
	w.PrivateKey = *priv
	w.PublicKey = append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	return nil
}
