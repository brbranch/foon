package foon

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"time"
)

type WriteBatch interface {
	Create(data interface{}) WriteBatch
	Set(data interface{}, opts ...firestore.SetOption) WriteBatch
	Delete(key *Key, opts ...firestore.Precondition) WriteBatch
	Commit() error
}

type WriteBatchImpl struct {
	context context.Context
	batch   *firestore.WriteBatch
	client *firestore.Client
	cache   *FirestoreCache
	logger  Logger
	updates []*KeyAndData
	deletes []*Key
	matadatas map[string]*Key
}

func (b *WriteBatchImpl) Create(data interface{}) WriteBatch {
	return b.put(data, func(doc *firestore.DocumentRef, data interface{}) {
		b.batch.Create(doc, data)
	})
}

func (b *WriteBatchImpl) Set(data interface{}, opts ...firestore.SetOption) WriteBatch {
	return b.put(data, func(doc *firestore.DocumentRef, data interface{}) {
		b.batch.Set(doc, data, opts...)
	})
	return b
}

func (b *WriteBatchImpl) put(data interface{}, fn func(doc *firestore.DocumentRef, data interface{})) WriteBatch {
	info, err := newFields(data)
	if err != nil {
		b.logger.Warning(fmt.Sprintf("failed to create Fields (reason: %v)", err))
		panic("invalid interface")
	}
	key := newKey(info)
	if info.HasUniqueID() {
		info.UpdateTime(time.Now())
		fn(key.CreateDocumentRef(b.client), data)
	} else {
		ref := key.CreateCollectionRef(b.client).NewDoc()
		info.UpdateField(ref.ID, time.Now())
		fn(ref, data)
	}

	b.matadatas[key.CollectionPath()] = key
	b.updates = append(b.updates, &KeyAndData{key, data})
	return b
}

func (b *WriteBatchImpl) Delete(key *Key, opts ...firestore.Precondition) WriteBatch {
	b.batch.Delete(key.CreateDocumentRef(b.client), opts...)
	b.deletes = append(b.deletes, key)
	b.matadatas[key.CollectionPath()] = key
	return b
}

func (b *WriteBatchImpl) Commit() error {
	if _ , err := b.batch.Commit(b.context); err != nil {
		return err
	}

	if len(b.updates) > 0 {
		b.cache.PutMulti(b.updates)
	}

	if len(b.deletes) > 0 {
		b.cache.DeleteMulti(b.deletes)
	}

	for _, key := range b.matadatas {
		metadata := LoadMetadata(b.cache, key)
		metadata.DeleteAll()
	}

	return nil
}
