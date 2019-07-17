package foon

import (
	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
	"google.golang.org/appengine"
	"google.golang.org/api/iterator"
	"context"
	"fmt"
	"errors"
	"reflect"
	"time"
)

type Foon struct {
	projectId string
	context.Context
	cache       *FirestoreCache
	client      FirestoreClient
	transaction bool
	logger      Logger
}

type KeyAndData struct {
	Key *Key
	Src interface{}
}

func New(ctx context.Context) (*Foon, error) {
	projectID := appengine.AppID(ctx)
	return NewStoreWithProjectID(ctx, projectID)
}

func Must(ctx context.Context) *Foon {
	res, err := New(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to create foon (reason: %+v)", err))
	}
	return res
}

func NewStoreWithProjectID(ctx context.Context, projectID string) (*Foon, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return nil, err
	}

	logger := &defaultLogger{ctx}
	client, err := app.Firestore(ctx)

	if err != nil {
		return nil, err
	}

	return &Foon{
		projectId:   projectID,
		Context:     ctx,
		client:      &FirestoreClientImpl{ctx, client},
		cache:       &FirestoreCache{ctx, logger},
		transaction: false,
		logger:      &defaultLogger{ctx},
	}, nil
}

func newStoreWithTransaction(foon *Foon, context context.Context, transaction *firestore.Transaction) *Foon {
	return &Foon{
		projectId:   foon.projectId,
		Context:     context,
		client:      &FirestoreTransactionClient{transaction, foon.client.Client()},
		cache:       foon.cache,
		transaction: true,
		logger:      foon.logger,
	}
}

func NotFound(err error) bool {
	return NoSuchDocument.Is(err)
}

func (s *Foon) SetLogger(logger Logger) {
	s.logger = logger
	s.cache.logger = logger
}

func (s *Foon) Put(src interface{}) error {
	info, err := newFields(src)
	if err != nil {
		return err
	}

	if info.HasUniqueID() {
		return s.put(info, src)
	}

	return s.insert(info, src)
}

func (s *Foon) Insert(src interface{}) error {
	info, err := newFields(src)
	if err != nil {
		return err
	}
	return s.insert(info, src)
}

func (s *Foon) InsertMulti(slices interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(slices))

	if value.Kind() != reflect.Slice {
		return errors.New("src must be slice pointer.")
	}

	batch, err := s.Batch()
	if err != nil {
		return err
	}

	length := value.Len()
	for i := 0; i < length; i++ {
		res := value.Index(i).Interface()
		batch.Create(res)
	}

	return batch.Commit()
}

func (s *Foon) PutMulti(slices interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(slices))

	if value.Kind() != reflect.Slice {
		return errors.New("src must be slice pointer.")
	}

	batch, err := s.Batch()
	if err != nil {
		return err
	}

	length := value.Len()
	for i := 0; i < length; i++ {
		res := value.Index(i).Interface()
		batch.Set(res)
	}

	return batch.Commit()
}

func (s *Foon) insert(info *fields, src interface{}) error {
	command := func(client FirestoreClient) error {
		key := newKey(info)
		col := key.CreateCollectionRef(client.Client())
		var ref *firestore.DocumentRef = nil
		if info.HasUniqueID() {
			ref = col.Doc(info.id.ID)
		} else {
			ref = col.NewDoc()
		}
		info.UpdateField(ref.ID, time.Now())

		s.logger.Trace(fmt.Sprintf("insert data (Path: %s, ID: %s)", ref.Path, ref.ID))

		_, err := ref.Create(s, src)
		if err != nil {
			return err
		}

		metadata := LoadMetadata(s.cache, key)
		metadata.DeleteAll()
		return nil
	}
	if err := s.execute(command); err != nil {
		s.warningf("failed to insert data (reason: %v)", err)
		return err
	}

	return s.setMemcache(info, src)
}

func (s *Foon) put(info *fields, src interface{}) error {
	err := s.execute(func(client FirestoreClient) error {
		key := newKey(info)
		ref := key.CreateDocumentRef(client.Client())
		s.logger.Trace(fmt.Sprintf("update data (Path: %s, ID: %s)", ref.Path, ref.ID))
		info.UpdateTime(time.Now())

		_, err := ref.Set(s, src)
		metadata := LoadMetadata(s.cache, key)
		metadata.DeleteAll()

		return err
	})

	if err != nil {
		return err
	}

	return s.setMemcache(info, src)
}

func (s *Foon) Get(src interface{}) error {
	info, err := newFields(src)
	if err != nil {
		return err
	}
	if !info.HasUniqueID() {
		return errors.New("Get method must be spesified ID")
	}

	if s.transaction {
		return s.getWithoutCache(info, src)
	}

	if err := s.cache.Get(newKey(info), src); err == nil {
		s.tracef("Get from Memcached.")
		return nil
	} else if !NoSuchDocument.Is(err) {
		s.warningf("failed to get Memcache %+v", err)
	}

	return s.getWithoutCache(info, src)
}

func (s *Foon) GetAll(key *Key, src interface{}) error {
	return s.GetByQuery(key, src, NoCondition)
}

func (s *Foon) GetMulti(src interface{}) error {
	if err := s.validSlice(src); err != nil {
		return err
	}

	if s.transaction {
		return s.GetMultiWithoutCache(src)
	}

	caches := map[string]*CacheResult{}

	original := reflect.Indirect(reflect.ValueOf(src))

	num := original.Len()

	for i := 0; i < num; i++ {
		s := original.Index(i).Interface()
		key := NewKey(s)
		if !key.HasUniqueID() {
			return errors.New("ID is required.")
		}
		cacheKey := InstanceCache.CreateURIByKey(key).URI()
		caches[cacheKey] = &CacheResult{
			Key:      key,
			Src:      s,
			HasCache: false,
		}
	}

	if err := s.cache.GetMulti(caches); err != nil {
		if NoSuchDocument.IsNot(err) {
			return err
		}
	}

	allHit := true

	for _, cache := range caches {
		if !cache.HasCache {
			allHit = false
			break
		}
	}

	if allHit {
		return nil
	}

	client := s.client

	refs := []*firestore.DocumentRef{}
	nonCaches := []*CacheResult{}
	for _, cache := range caches {
		if !cache.HasCache {
			refs = append(refs, cache.Key.CreateDocumentRef(client.Client()))
			nonCaches = append(nonCaches, &CacheResult{cache.Key, cache.Src, false})
		}
	}

	values, err := client.GetAll(refs)
	if err != nil {
		return err
	}

	if len(values) != len(refs) {
		// すべて取得できてたら同じになるはず
		s.logger.Warning("invalid data")
		return NoSuchDocument
	}

	results := []*KeyAndData{}
	for _, doc := range values {
		for _, cache := range nonCaches {
			if cache.Key.SamePath(doc.Ref) {
				if err := doc.DataTo(cache.Src); err != nil {
					return err
				}
				results = append(results, &KeyAndData{cache.Key, cache.Src})
			}
		}
	}

	s.cache.PutMulti(results)
	return nil
}

func (s *Foon) GetMultiWithoutCache(src interface{}) error {
	if err := s.validSlice(src); err != nil {
		return err
	}
	client := s.client
	original := reflect.Indirect(reflect.ValueOf(src))

	num := original.Len()
	refs := []*firestore.DocumentRef{}

	for i := 0; i < num; i++ {
		s := original.Index(i).Interface()
		key := NewKey(s)
		if !key.HasUniqueID() {
			return errors.New("ID is required.")
		}
		refs = append(refs, key.CreateDocumentRef(client.Client()))
	}

	values, err := client.GetAll(refs)
	if err != nil {
		return err
	}
	sliceType := reflect.TypeOf(original.Interface())
	slices := reflect.MakeSlice(sliceType, 0, 0)

	results := []*KeyAndData{}

	for _, doc := range values {
		src := reflect.New(original.Type().Elem()).Interface()
		if err := doc.DataTo(src); err != nil {
			return err
		}
		newValue := reflect.Indirect(reflect.ValueOf(src))
		newSrc := newValue.Interface()

		key := NewKey(newSrc)
		results = append(results, &KeyAndData{key, newSrc})
		slices = reflect.Append(slices, newValue)
	}

	if original.Len() != slices.Len() {
		// すべて取得できてたら同じになるはず
		s.logger.Warning("invalid data")
		return NoSuchDocument
	}

	original.Set(slices)
	s.cache.PutMulti(results)

	return nil
}

func (s *Foon) GetByQueryWithoutCache(key *Key, src interface{}, conditions *Conditions) error {
	if err := s.validSlice(src); err != nil {
		return err
	}

	return s.getChildrenWithoutCache(key, src, conditions)
}

func (s *Foon) GetByQuery(key *Key, src interface{}, conditions *Conditions) error {
	if err := s.validSlice(src); err != nil {
		return err
	}

	if s.transaction {
		return s.getChildrenWithoutCache(key, src, conditions)
	}

	metadata := LoadMetadata(s.cache, key)
	if err := metadata.Load(conditions.URI(key), src); err == nil {
		s.logger.Trace("cache is hit! ")
		return nil
	}

	return s.getChildrenWithoutCache(key, src, conditions)
}

func (s *Foon) getChildrenWithoutCache(parentKey *Key, slices interface{}, conditions *Conditions) error {
	value := reflect.Indirect(reflect.ValueOf(slices))

	col := parentKey.CreateCollectionRef(s.client.Client())
	it := conditions.Query(col).Documents(s)

	for {
		doc, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}

		src := reflect.New(value.Type().Elem()).Interface()
		if err := doc.DataTo(src); err != nil {
			return err
		}

		value.Set(reflect.Append(value, reflect.Indirect(reflect.ValueOf(src))))
	}

	dst := value.Interface()
	meta := LoadMetadata(s.cache, parentKey)
	meta.Put(conditions.URI(parentKey), dst)

	return nil
}

func (s *Foon) GetWithoutCache(src interface{}) error {
	info, err := newFields(src)
	if err != nil {
		return err
	}
	if !info.HasUniqueID() {
		return errors.New("Get method must be spesified ID")
	}

	return s.getWithoutCache(info, src)
}

func (s *Foon) RunInTransaction(fn func(f *Foon) error, options ...firestore.TransactionOption) error {
	return s.client.RunTransaction(func(ctx context.Context, fs *firestore.Transaction) error {
		newFoon := newStoreWithTransaction(s, ctx, fs)
		return fn(newFoon)
	}, options...)
}

func (s *Foon) Batch() (WriteBatch, error) {
	batch, err := s.client.Batch()
	if err != nil {
		return nil, err
	}
	return &WriteBatchImpl{
		context:   s.Context,
		batch:     batch,
		client:    s.client.Client(),
		cache:     s.cache,
		logger:    s.logger,
		updates:   []*KeyAndData{},
		deletes:   []*Key{},
		matadatas: map[string]*Key{},
	}, nil
}

func (s *Foon) getWithoutCache(info *fields, src interface{}) error {
	err := s.execute(func(client FirestoreClient) error {
		key := newKey(info)
		docRef := key.CreateDocumentRef(client.Client())
		s.logger.Trace(fmt.Sprintf("try to get firestore (path: %s)", docRef.Path))
		doc, err := docRef.Get(s)
		if err != nil {
			if doc != nil && doc.Exists() == false {
				return NoSuchDocument
			}
			return err
		}
		info.updateKey(docRef)

		return doc.DataTo(src)
	})
	if err != nil {
		return err
	}
	return s.setMemcache(info, src)
}

func (s *Foon) execute(fn func(client FirestoreClient) error) error {

	if err := fn(s.client); err != nil {
		return err
	}

	return nil
}

func (s *Foon) setMemcache(info *fields, src interface{}) error {
	if err := s.cache.Put(newKey(info), src); err != nil {
		s.warningf("failed to Put Memcached %+v", err)
		return err
	}
	return nil
}

func (s *Foon) setMemcacheMulti(res []*KeyAndData) error {
	return s.cache.PutMulti(res)
}

func (s *Foon) setMemcacheWithKey(key string, src interface{}) error {
	if err := s.cache.PutCache(key, src); err != nil {
		s.warningf("failed to Put Memcached %+v", err)
		return err
	}
	return nil
}

func (s *Foon) Delete(src interface{}) error {
	key := NewKey(src)
	s.cache.Delete(key)
	return s.client.Delete(key.CreateDocumentRef(s.client.Client()))
}

func (s *Foon) tracef(format string, args ...interface{}) {
	tracef(s.logger, format, args...)
}

func (s *Foon) warningf(format string, args ...interface{}) {
	warningf(s.logger, format, args...)
}

func (s *Foon) validSlice(src interface{}) error {
	if reflect.ValueOf(src).Kind() != reflect.Ptr {
		return errors.New("src must be slice pointer.")
	}

	value := reflect.Indirect(reflect.ValueOf(src))

	if value.Kind() != reflect.Slice {
		return errors.New("src must be slice.")
	}

	return nil
}
