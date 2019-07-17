package foon

import "fmt"

type CacheKey string

const (
	InstanceCache CacheKey = "foon"
	CollectionCache CacheKey = "foon/collections"
	MetadataCache CacheKey = "foon/metadata"
)

type CacheURI string

func (c CacheURI) URI() string {
	return string(c)
}

func (c CacheKey) CreateURIByKey(key *Key) IURI {
	if c == InstanceCache {
		return CacheURI(fmt.Sprintf("%s/%s", c, key.Path()))
	}
	return CacheURI(fmt.Sprintf("%s/%s", c, key.CollectionPath()))
}


