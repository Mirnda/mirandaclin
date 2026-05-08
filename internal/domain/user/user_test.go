package user

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	var password = "mirandaClin"
	salt1, hash1, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword erro: %v", err)
	}

	salt2, hash2, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword erro: %v", err)
	}

	if salt1 == salt2 {
		t.Error("salt deve ser diferente a cada chamada")
	}
	if hash1 == hash2 {
		t.Error("hash deve ser diferente a cada chamada")
	}
}
