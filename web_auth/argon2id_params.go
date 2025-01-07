package web_auth

import "github.com/Morpheus0x/argon2id"

var argonParams = argon2id.Params{
	Memory:      19 * 1024,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}
