package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash           = errors.New("invalid password hash")
	ErrUnsupportedAlgorithm  = errors.New("unsupported password algorithm")
	ErrUnsupportedVersion    = errors.New("unsupported hash version")
	ErrUnsupportedParameters = errors.New("unsupported password hash parameters")
)

const (
	memory  uint32 = 64 * 1024
	time    uint32 = 3
	threads uint8  = 1
	keyLen  uint32 = 32
	saltLen        = 16
)

func Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		keyLen,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	result := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		time,
		threads,
		encodedSalt,
		encodedHash,
	)

	return result, nil
}

func Verify(password string, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" {
		return false, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return false, ErrUnsupportedAlgorithm
	}

	key, value, ok := strings.Cut(parts[2], "=")
	if !ok || key != "v" {
		return false, ErrInvalidHash
	}
	version, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return false, ErrInvalidHash
	}
	if version != argon2.Version {
		return false, ErrUnsupportedVersion
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, ErrInvalidHash
	}

	pMemory, err := parseParameter(params[0], "m", 32)
	if err != nil {
		return false, err
	}
	pTime, err := parseParameter(params[1], "t", 32)
	if err != nil {
		return false, err
	}
	pThreads, err := parseParameter(params[2], "p", 8)
	if err != nil {
		return false, err
	}
	if pMemory != memory ||
		pTime != time ||
		uint8(pThreads) != threads {
		return false, ErrUnsupportedParameters
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != saltLen {
		return false, ErrInvalidHash
	}
	savedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || uint32(len(savedHash)) != keyLen {
		return false, ErrInvalidHash
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		pTime,
		pMemory,
		uint8(pThreads),
		keyLen,
	)

	if subtle.ConstantTimeCompare(actualHash, savedHash) != 1 {
		return false, nil
	}

	return true, nil
}

func parseParameter(param string, key string, bitSize int) (uint32, error) {
	pKey, value, ok := strings.Cut(param, "=")
	if !ok || pKey != key {
		return 0, ErrInvalidHash
	}
	result, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return 0, ErrInvalidHash
	}

	return uint32(result), nil
}
