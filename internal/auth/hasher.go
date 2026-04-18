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
}

func (h *Argon2Hasher) GetName() string {
	return ARGON2id
}

func NewArgon2Hasher(memory uint32, iterations uint32, parallelism uint8, keyLength uint32) *Argon2Hasher {
	return &Argon2Hasher{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		keyLength:   keyLength,
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
	hash := h.hashInternal(arHash.Version, uint8(arHash.Parallelism), uint32(arHash.Iterations),
		uint32(arHash.Memory), h.keyLength, password, arHash.Salt)

	if !hash.Equals(arHash) {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func (h *Argon2Hasher) Hash(password string, salt []byte) (string, error) {
	arHash := h.hashInternal(argon2.Version, h.parallelism, h.iterations, h.memory, h.keyLength, password, salt)
	return arHash.ToArgon2String(), nil
}

func (h *Argon2Hasher) hashInternal(version int, parallelism uint8, iterations, memory, keyLength uint32, password string, salt []byte) *Argon2Hash {
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
	return &argonHash
}
