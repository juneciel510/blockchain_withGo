package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ripemd160"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

// newKeyPair creates a new cryptographic key pair
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	publicKey:=privateKey.PublicKey
	return *privateKey, pubKeyToByte(publicKey)
}

// pubKeyToByte converts the ecdsa.PublicKey to a concatenation of its coordinates in bytes
func pubKeyToByte(pubkey ecdsa.PublicKey) []byte {
	// step 1 of: https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
	pubKeyByte:=[]byte{}
	pubKeyByte = append(pubKeyByte, pubkey.X.Bytes()...)
	pubKeyByte = append(pubKeyByte, pubkey.Y.Bytes()...)
	return pubKeyByte
}

// GetAddress returns address
// https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
func GetAddress(pubKeyBytes []byte) []byte {
	hashRipemd160:=HashPubKey(pubKeyBytes)
	version:=[]byte{version}
	versionedPayload:=append(version,hashRipemd160...)
	checksum:=checksum(versionedPayload)
	appendChecksum:=append(versionedPayload,checksum...)
	address:=Base58Encode(appendChecksum)
	return address
}

// GetStringAddress returns address as string
func GetStringAddress(pubKeyBytes []byte) string {
	return string(pubKeyBytes)
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	// compute the SHA256 + RIPEMD160 hash of the pubkey
	// step 2 and 3 of: https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
	// use the go package ripemd160:
	// https://godoc.org/golang.org/x/crypto/ripemd160
	h:=sha256.New()
	h.Write(pubKey)
	hashSha256:=h.Sum(nil)
	h=ripemd160.New()
	h.Write(hashSha256)
	hashRipemd160:=h.Sum(nil)
	return hashRipemd160
}

// GetPubKeyHashFromAddress returns the hash of the public key
// discarding the version and the checksum
func GetPubKeyHashFromAddress(address string) []byte {
	addressDecoded:=Base58Decode([]byte(address))
	length := len(addressDecoded)
	pubKeyHash:=addressDecoded[1:length-addressChecksumLen]
	return pubKeyHash

}

// ValidateAddress check if an address is valid
func ValidateAddress(address string) bool {
	// Validate a address by decoding it, extracting the
	// checksum, re-computing it using the "checksum" function
	// and comparing both.
	addressDecoded:=Base58Decode([]byte(address))
	length := len(addressDecoded)
	pubKeyHash:=addressDecoded[1:length-addressChecksumLen]
	checksumExtracted:=addressDecoded[length-addressChecksumLen:length]
	version:=[]byte{version}
	versionedPayload:=append(version,pubKeyHash...)
	return bytes.Equal(checksumExtracted,checksum(versionedPayload))
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	// Perform a double sha256 on the versioned payload
	// and return the first 4 bytes
	// Steps 5,6, and 7 of: https://en.bitcoin.it/wiki/Technical_background_of_version_1_Bitcoin_addresses#How_to_create_Bitcoin_Address
	h:=sha256.New()
	h.Write(payload)
	hashVP:=h.Sum(nil)
	h=sha256.New()
	h.Write(hashVP)
	hashHashVP:=h.Sum(nil)
	checksum:=hashHashVP[0:4]
	return checksum
}

func encodeKeyPair(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	return encodePrivateKey(privateKey), encodePublicKey(publicKey)
}

func encodePrivateKey(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

func encodePublicKey(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncodedPub)
}

func decodeKeyPair(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	return decodePrivateKey(pemEncoded), decodePublicKey(pemEncodedPub)
}

func decodePrivateKey(pemEncoded string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(pemEncoded))
	privateKey, _ := x509.ParseECPrivateKey(block.Bytes)

	return privateKey
}

func decodePublicKey(pemEncodedPub string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	genericPubKey, _ := x509.ParsePKIXPublicKey(blockPub.Bytes)
	publicKey := genericPubKey.(*ecdsa.PublicKey) // cast to ecdsa

	return publicKey
}
