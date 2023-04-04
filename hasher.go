package imagine

import (
	"crypto/md5"
	"encoding/hex"
)

type Hasher interface {
	Hash([]byte) (string, error)
}

func MD5Hasher() Hasher {
	return &md5Hasher{}
}

type md5Hasher struct{}

func (m *md5Hasher) Hash(b []byte) (string, error) {
	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:]), nil
}
