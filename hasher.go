package imagine

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// Hasher describes the hashing algorithm used to generate the image name.
// The default implementation is MD5Hasher.
// You can implement your own hashing algorithm by implementing this interface.
// For example, you can use SHA256 instead of MD5.
// You can also use a combination of hashing algorithms.
type Hasher interface {
	Hash([]byte) (string, error)
}

// MD5Hasher is the default implementation of Hasher.
// It uses MD5 to generate the image name.
func MD5Hasher() Hasher {
	return &md5Hasher{}
}

type md5Hasher struct{}

// Hash generates the image name using MD5.
func (m *md5Hasher) Hash(b []byte) (string, error) {
	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:]), nil
}

// Sha256Hasher is the SHA256 implementation of Hasher.
func SHA256Hasher() Hasher {
    return &sha256Hasher{}
}

// implement SHA256 for Hasher interface
type sha256Hasher struct{}

func (s *sha256Hasher) Hash(b []byte) (string, error) {
    hash := sha256.Sum256(b)
    return hex.EncodeToString(hash[:]), nil
}


