package internal

import "github.com/itering/substrate-api-rpc/metadata"

type storage struct {
	specVersion map[int]int
	hash        map[int]string
	metadata    map[int]*metadata.Instant
}

var GlobalStorage storage

func (s *storage) PutHash(height int, hash string) {
	s.hash[height] = hash
}

func (s *storage) GetHash(height int) string {
	return s.hash[height]
}

func (s *storage) PutSpecVersion(height, version int) {
	s.specVersion[height] = version
}

func (s *storage) GetSpecVersion(height int) int {
	return s.specVersion[height]
}

func (s *storage) PutMetadata(height int, meta *metadata.Instant) {
	s.metadata[height] = meta
}

func (s *storage) GetMetadata(height int) *metadata.Instant {
	return s.metadata[height]
}

func (s *storage) DeleteBlockInfo(height int) {
	delete(s.specVersion, height)
	delete(s.hash, height)
	delete(s.metadata, height)
}
