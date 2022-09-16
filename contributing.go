package skiperator

import (
	_ "embed"
	"hash/fnv"
)

//go:embed CONTRIBUTING.md
var contributing string
var Checksum uint64

func init() {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(contributing))
	Checksum = hash.Sum64()
}
