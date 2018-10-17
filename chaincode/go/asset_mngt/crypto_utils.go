package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// ==== In cryptography, a keyed-hash message authentication code (HMAC)
// ==== is a specific type of message authentication code (MAC) involving
// ==== a cryptographic hash function and a secret cryptographic key.
func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// 如果messageMAC是message的合法HMAC标签，函数返回真
func CheckMAC(message string, messageMAC string, secret string) bool {
	key := []byte(secret)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message))
	expectedMACBytes := mac.Sum(nil)
	checkBytes, err := Base64Decoding(messageMAC)
	if err != nil {
		return false
	}

	return hmac.Equal(checkBytes, expectedMACBytes)
}
func ComputeSHA256Bytes(message string) []byte {
	h := sha256.New()
	h.Write([]byte(message))
	return h.Sum(nil)
}

func ComputeSHA256Base64(message string) string {
	return Base64Encoding(ComputeSHA256Bytes(message))
}

func ComputeSHA256Base16UpperCase(message string) string {
	return fmt.Sprintf("%X", ComputeSHA256Bytes(message))
}

func ComputeSHA256Base16LowerCase(message string) string {
	return fmt.Sprintf("%x", ComputeSHA256Bytes(message))
}

func CheckSHA256(message string, messageMAC string) bool {
	h := sha256.New()
	h.Write([]byte(message))
	expectedSHA256Bytes := h.Sum(nil)
	checkBytes, err := Base64Decoding(messageMAC)
	if err != nil {
		return false
	}

	if len(checkBytes) != len(expectedSHA256Bytes) {
		return false
	}

	for i := 0; i < len(checkBytes); i++ {
		if checkBytes[i] != expectedSHA256Bytes[i] {
			return false
		}
	}

	return true
}

func Base64Encoding(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func Base64Decoding(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
