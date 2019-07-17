package foon

import (
	"fmt"
	"google.golang.org/appengine/memcache"
	"time"
)

type IURI interface {
	URI() string
}

/** クエリ結果などを保持するためのキャッシュ (関連のPut時には全削除される) */
type CacheMetadata struct {
	Item  *CacheMetadataItem
	cache *FirestoreCache
}

type CacheMetadataItem struct {
	MemcachePath string
	Data         []string
}

type MetadataItem struct {
	Key  IURI
	Data interface{}
}

func LoadMetadata(cache *FirestoreCache, key *Key) *CacheMetadata {
	path := MetadataCache.CreateURIByKey(key).URI()
	res := &CacheMetadataItem{}

	cache.logger.Trace(fmt.Sprintf("load metadata (key: %s)", path))

	err := cache.GetCache(path, res)
	if err == nil {
		cache.logger.Trace(fmt.Sprintf("metadata cache is hit (%+v)", res.Data))
		return &CacheMetadata{res, cache}
	} else {
		if NoSuchDocument.IsNot(err) {
			cache.logger.Warning(fmt.Sprintf("failed to load metadata (reason: %+v)", err))
		}
	}

	res.MemcachePath = path
	res.Data = []string{}
	return &CacheMetadata{res, cache}
}

func (c *CacheMetadata) Save() error {
	return c.cache.PutCache(c.Item.MemcachePath, c.Item)
}

func (c *CacheMetadata) DeleteAll() error {
	c.cache.logger.Trace(fmt.Sprintf("delete metadata (%s)", c.Item.MemcachePath))
	if len(c.Item.Data) == 0 {
		return nil
	}
	keys := []string{}
	for _, path := range c.Item.Data {
		keys = append(keys, path)
	}
	keys = append(keys, c.Item.MemcachePath)
	c.Item.Data = []string{}
	return memcache.DeleteMulti(c.cache.Context, keys)
}

func (c *CacheMetadata) Has(key IURI) bool {
	for _, name := range c.Item.Data {
		if name == key.URI() {
			return true
		}
	}
	return false
}

func (c *CacheMetadata) Put(uri IURI, src interface{}) error {
	return c.PutMulti([]MetadataItem{MetadataItem{uri, src}})
}

func (c *CacheMetadata) Load(key IURI, src interface{}) error {
	c.cache.logger.Trace("try to load Cache.")
	if !c.Has(key) {
		c.cache.logger.Trace("cache is not registy")
		return NoSuchDocument
	}

	return c.cache.GetCache(key.URI(), src)
}

func (c *CacheMetadata) PutMulti(datas []MetadataItem) error {
	strs := []string{}
	items := []*memcache.Item{}
	for _, data := range datas {
		if !c.Has(data.Key) {
			c.Item.Data = append(c.Item.Data, data.Key.URI())
		}
		bytes, err := c.cache.asByte(data.Data)
		if err != nil {
			c.cache.logger.Warning(fmt.Sprintf("failed to save cache. (reason: %v)", err))
			return err
		}
		strs = append(strs, data.Key.URI())
		items = append(items, &memcache.Item{
			Key:        data.Key.URI(),
			Value:      bytes,
			Expiration: time.Hour * 24 * 5,
		})
	}
	bytes, err := c.cache.asByte(c.Item)
	if err != nil {
		c.cache.logger.Warning(fmt.Sprintf("failed to save metadata. (reason: %v)", err))
		return err
	}
	strs = append(strs, c.Item.MemcachePath)
	items = append(items, &memcache.Item{
		Key:        c.Item.MemcachePath,
		Value:      bytes,
		Expiration: time.Hour * 24 * 5,
	})

	c.cache.logger.Trace(fmt.Sprintf("metadata save (%+v)", strs))

	err = memcache.SetMulti(c.cache.Context, items)
	if err != nil {
		c.cache.logger.Warning(fmt.Sprintf("failed to save cache (reason: %v)", err))
	}

	return err
}
