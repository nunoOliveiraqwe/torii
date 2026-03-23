package auth

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

type Argon2Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
	version     int
}

func (h *Argon2Hasher) GetName() string {
	return ARGON2id
}

func NewArgon2Hasher(version int, memory uint32, iterations uint32, parallelism uint8, keyLength uint32) *Argon2Hasher {
	return &Argon2Hasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		keyLength:   keyLength,
		version:     version,
	}
}

func (h *Argon2Hasher) CompareHashAndPassword(hashedPwd string, password string) error {
	var arHash *Argon2Hash
	var err error
	arHash, err = FromArgon2String(hashedPwd)
	if err != nil {
		zap.S().Errorf("error decoding hashed pwd, %v", err)
		return err
	}
	hash, err := h.hashInternal(arHash.Version, uint8(arHash.Parallelism), uint32(arHash.Iterations),
		uint32(arHash.Memory), h.keyLength, password, arHash.Salt)
	if err != nil {
		zap.S().Errorf("error hashing password, %v", err)
		return err
	}
	if !hash.Equals(arHash) {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func (h *Argon2Hasher) getArgon2Params() Argon2Params {
	params := Argon2Params{
		Memory:      h.memory,
		Iterations:  h.iterations,
		Parallelism: h.parallelism,
		KeyLength:   h.keyLength,
	}
	return params
}

func (h *Argon2Hasher) Hash(password string, salt []byte) (string, error) {
	arHash, err := h.hashInternal(h.version, h.parallelism, h.iterations, h.memory, h.keyLength, password, salt)
	if err != nil {
		zap.S().Errorf("error hashing password, %v", err)
		return "", err
	}
	return arHash.ToArgon2String(), nil
}

func (h *Argon2Hasher) hashInternal(version int, parallelism uint8, iterations, memory, keyLength uint32, password string, salt []byte) (*Argon2Hash, error) {
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	argonHash := Argon2Hash{
		AlgorithmName: ARGON2id,
		Version:       version,
		HashedData:    hash,
		Salt:          salt,
		Iterations:    int(iterations),
		Memory:        int(memory),
		Parallelism:   int(parallelism),
	}
	return &argonHash, nil
}
