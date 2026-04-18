package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Argon2Hash struct {
	AlgorithmName string
	Version       int
	HashedData    []byte
	Salt          []byte
	Iterations    int
	Parallelism   int
	Memory        int
}

const (
	ARGON2i  = "argon2i"
	ARGON2d  = "argon2d"
	ARGON2id = "argon2id"
)

var (
	algo_argon2 = map[string]struct{}{
		ARGON2i:  {},
		ARGON2d:  {},
		ARGON2id: {},
	}
)

func FromArgon2String(input string) (*Argon2Hash, error) {
	if !strings.HasPrefix(input, "$") {
		return nil, fmt.Errorf("unsupported input: %s", input)
	}
	parts := strings.Split(input[1:], "$")
	if len(parts) < 5 {
		return nil, errors.New("invalid input format")
	}
	return argon2HashFromString(parts)
}

func argon2HashFromString(parts []string) (*Argon2Hash, error) {
	algorithmName := strings.TrimSpace(parts[0])
	if _, exists := algo_argon2[algorithmName]; !exists {
		return nil, fmt.Errorf("unsupported algorithm: %s. Expected one of: %v", algorithmName, getAlgorithmNames())
	}

	if !strings.HasPrefix(parts[1], "v=") {
		return nil, fmt.Errorf("missing version prefix 'v=' in: %s", parts[1])
	}
	version, err := strconv.Atoi(parts[1][2:])
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", parts[1])
	}

	parameters := parts[2]
	memoryKiB, err := parseMemory(parameters)
	if err != nil {
		return nil, fmt.Errorf("invalid memory parameter: %v", err)
	}

	iterations, err := parseIterations(parameters)
	if err != nil {
		return nil, fmt.Errorf("invalid iterations parameter: %v", err)
	}

	parallelism, err := parseParallelism(parameters)
	if err != nil {
		return nil, fmt.Errorf("invalid parallelism parameter: %v", err)
	}

	salt, err := base64.StdEncoding.WithPadding(-1).DecodeString(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid salt: %v", err)
	}

	hashedData, err := base64.StdEncoding.WithPadding(-1).DecodeString(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid hashed data: %v", err)
	}

	return &Argon2Hash{
		AlgorithmName: algorithmName,
		Version:       version,
		HashedData:    hashedData,
		Salt:          salt,
		Iterations:    iterations,
		Memory:        memoryKiB,
		Parallelism:   parallelism,
	}, nil
}

func (h *Argon2Hash) ToArgon2String() string {
	parameters := fmt.Sprintf("t=%d,m=%d,p=%d", h.Iterations, h.Memory, h.Parallelism)
	str := fmt.Sprintf("$%s$v=%d$%s$%s$%s",
		h.AlgorithmName,
		h.Version,
		parameters,
		base64.StdEncoding.WithPadding(-1).EncodeToString(h.Salt),
		base64.StdEncoding.WithPadding(-1).EncodeToString(h.HashedData),
	)
	return str
}

func (h *Argon2Hash) Equals(other *Argon2Hash) bool {
	if other == nil {
		return false
	}
	if h.Version != other.Version {
		return false
	}
	if h.AlgorithmName != other.AlgorithmName {
		return false
	}
	if h.Iterations != other.Iterations {
		return false
	}
	if h.Parallelism != other.Parallelism {
		return false
	}
	if h.Memory != other.Memory {
		return false
	}

	return subtle.ConstantTimeCompare(h.HashedData, other.HashedData) == 1 &&
		subtle.ConstantTimeCompare(h.Salt, other.Salt) == 1
}

func getAlgorithmNames() []string {
	var names []string
	for name := range algo_argon2 {
		names = append(names, name)
	}
	return names
}

func parseMemory(parameters string) (int, error) {
	return parseParameter(parameters, "m=")
}

func parseIterations(parameters string) (int, error) {
	return parseParameter(parameters, "t=")
}

func parseParallelism(parameters string) (int, error) {
	return parseParameter(parameters, "p=")
}

func parseParameter(parameters, prefix string) (int, error) {
	parts := strings.Split(parameters, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, prefix) {
			return strconv.Atoi(part[len(prefix):])
		}
	}
	return 0, fmt.Errorf("missing parameter %s in: %s", prefix, parameters)
}
