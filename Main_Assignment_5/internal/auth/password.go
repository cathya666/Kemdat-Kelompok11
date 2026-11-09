package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

var params = struct {
	memory, iterations    uint32
	parallelism           uint8
	saltLength, keyLength uint32
}{
	memory:      64 * 1024,
	iterations:  1,
	parallelism: 4,
	saltLength:  16,
	keyLength:   32,
}

func HashPassword(password string) (string, error) {
	var salt []byte = make([]byte, params.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	var hash []byte = argon2.IDKey(
		[]byte(password), salt,
		params.iterations, params.memory,
		params.parallelism, params.keyLength)

	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(hash), nil

}

func CheckPassword(password string, encoded_hash string) (bool, error) {
	parts := strings.Split(encoded_hash, "$")
	if len(parts) != 2 {
		return false, errors.New("Invalid Hash format")
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	expectedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	var hash []byte = argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)
	if subtle.ConstantTimeCompare(hash, expectedHash) == 1 {
		return true, nil
	}
	return false, nil
}
