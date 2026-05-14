package config

import (
	nob "github.com/Fipaan/nob.go"
)

const (
	defaultTest = "DEFAULT TEST"
)

func Test() string {
	return nob.TryEnv("TEST", defaultTest)
}
