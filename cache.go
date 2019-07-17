package foon

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"google.golang.org/appengine/memcache"
	"time"
	)

/** Memcacheを扱う */
type FirestoreCache struct {
	context.Context
	logger Logger
}

/** キャッシュを取得する際の結果 */
type CacheResult struct {
	Key      *Key
	Src      interface{}
	HasCache bool
}

func NewCache(ctx context.Context, logger Logger) *FirestoreCache {
	return &FirestoreCache{ctx, logger}
}

func (c *FirestoreCache) GetEntity(src interface{}) error {
	info, err := KeyError(src)
	if err != nil {
		return err
	}
	return c.Get(info, src)
}

func (c *FirestoreCache) Get(info *Key, src interface{}) error {
	if info.HasUniqueID() == false {
		return InvalidId
	}
	return c.GetCache(InstanceCache.CreateURIByKey(info).URI(), src)
}

func (c *FirestoreCache) GetCache(path string, src interface{}) error {
	c.logger.Trace(fmt.Sprintf("try to get memcache (path: %s)", path))
	if cache, err := memcache.Get(c, path); err == nil && cache != nil {
		c.logger.Trace(fmt.Sprintf("cache is hit (path: %s)", path))
		err := c.asValue(cache.Value, src)
		if err != nil {
			c.logger.Warning(fmt.Sprintf("failed to get cache (reason: %v)", err))
		}
		return err
	}
	return NoSuchDocument
}

func (c *FirestoreCache) GetMulti(results map[string]*CacheResult) error {
	c.logger.Trace(fmt.Sprintf("try to get memcaches"))
	keys := []string{}
	for key, val := range results {
		keys = append(keys, key)
		val.HasCache = false
	}

	if caches, err := memcache.GetMulti(c, keys); err == nil {
		for _, item := range caches {
			if m, ok := results[item.Key]; ok {
				c.logger.Trace(fmt.Sprintf("cache is hit (%s)", item.Key))
				err := c.asValue(item.Value, m.Src)
				if err != nil {
					c.logger.Warning(fmt.Sprintf("failed to get cache (reason: %v)", err))
					return err
				}
				m.HasCache = true
			}
		}
		return nil
	}
	return NoSuchDocument
}

func (c *FirestoreCache) PutEntity(src interface{}) error {
	info, err := KeyError(src)
	if err != nil {
		return err
	}
	return c.Put(info, src)
}

func (c *FirestoreCache) Put(info *Key, src interface{}) error {
	if info.HasUniqueID() == false {
		return InvalidId
	}
	return c.PutCache(InstanceCache.CreateURIByKey(info).URI(), src)
}

func (c *FirestoreCache) PutMulti(results []*KeyAndData) error {
	items := []*memcache.Item{}

	for _, res := range results {
		bytes, err := c.asByte(res.Src)
		if err != nil {
			return err
		}
		items = append(items, &memcache.Item{
			Key:        InstanceCache.CreateURIByKey(res.Key).URI(),
			Value:      bytes,
			Expiration: time.Hour * 24 * 5,
		})
	}

	if len(items) > 0 {
		return memcache.SetMulti(c, items)
	}

	return nil
}

func (c *FirestoreCache) PutCache(path string, src interface{}) error {
	bytes, err := c.asByte(src)
	if err != nil {
		return err
	}
	tracef(c.logger, "save to memcache (key: %s)", path)

	return memcache.Set(c, &memcache.Item{
		Key:        path,
		Value:      bytes,
		Expiration: time.Hour * 24 * 5,
	})
}

func (c *FirestoreCache) Delete(info *Key) error {
	if info.HasUniqueID() == false {
		return nil
	}
	url := InstanceCache.CreateURIByKey(info).URI()
	c.logger.Trace(fmt.Sprintf("delete cache (key: %s)", url))
	return memcache.Delete(c, url)
}

func (c *FirestoreCache) DeleteMulti(keys []*Key) error {
	deleteKeys := []string{}
	for _, key := range keys {
		deleteKeys = append(deleteKeys, InstanceCache.CreateURIByKey(key).URI())
	}
	return memcache.DeleteMulti(c, deleteKeys)
}

func (c *FirestoreCache) DeleteCache(path string) error {
	return memcache.Delete(c, path)
}

func (c *FirestoreCache) asByte(src interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(src)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *FirestoreCache) asValue(data []byte, src interface{}) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(src)
}
