package foon

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"regexp"
	"reflect"
)

type Key struct {
	ParentPath string
	Collection string
	ID string
}

func newKey(fields *fields) *Key {
	res := &Key{}
	if fields.self != nil {
		fields.self = res
	}
	if fields.parent != nil {
		res.ParentPath = fields.parent.parent.Path()
	}
	res.Collection = fields.CollectionName()
	if fields.id != nil {
		res.ID = fields.id.ID
	}
	return res
}

func newGroupKey(src interface{}) *Key {
	if info, ok := src.(*fields); ok {
		return &Key{
			ParentPath: "",
			Collection: info.CollectionName(),
			ID: "",
		}
	}
	info, err := newFields(src)
	if err != nil {
		panic(fmt.Sprintf("error is occurred (reason: %+v)" , err))
	}
	return &Key{
		ParentPath: "",
		Collection: info.CollectionName(),
		ID: "",
	}
}

func NewKey(src interface{}) *Key {
	if info, ok := src.(*fields); ok {
		return newKey(info)
	}
	info, err := newFields(src)
	if err != nil {
		panic(fmt.Sprintf("error is occurred (reason: %+v)" , err))
	}
	return newKey(info)
}

func KeyError(src interface{}) (*Key, error) {
	info, err := newFields(src)
	if err != nil {
		return nil ,err
	}
	return newKey(info), nil
}

func NewKeyWithPath(fullPath string) *Key {
	k := &Key{}
	k.Update(fullPath)
	return k
}

func (k *Key) SamePath(ref *firestore.DocumentRef) bool {
	other := NewKeyWithPath(ref.Path)
	return k.Equals(other)
}

func (k *Key) ParentKey() *Key {
	if k.ParentPath == "" {
		return nil
	}
	r := regexp.MustCompile(`^(.*)?/([^/]+)/([^/]+)$`)
	if r.MatchString(k.ParentPath) {
		res := &Key{}
		matchs := r.FindAllStringSubmatch(k.ParentPath, -1)
		res.ParentPath = matchs[0][1]
		res.Collection = matchs[0][2]
		res.ID = matchs[0][3]
		return res
	}
	r = regexp.MustCompile(`^([^/]+)/([^/]+)$`)
	if r.MatchString(k.ParentPath) {
		res := &Key{}
		matchs := r.FindAllStringSubmatch(k.ParentPath, -1)
		res.ParentPath = ""
		res.Collection = matchs[0][1]
		res.ID = matchs[0][2]
		return res
	}
	return nil
}

func (k *Key) Inject(src interface{}) error {
	if info, ok := src.(*fields); ok {
		 k.injectFields(info)
		 return nil
	}
	info, err := newFields(src)
	if err != nil {
		return err
	}
	k.injectFields(info)
	return nil
}

func (k *Key) injectFields(field *fields) {
	field.self = k
	if field.parent != nil {
		key := k.ParentKey()
		field.parent.parent = key
		field.parent.field.Set(reflect.ValueOf(key))
	}
	if field.id != nil {
		field.id.SetID(k.ID)
	}
}

func (k *Key) IsSameKind(src interface{}) bool {
	key := NewKey(src)
	return key.Collection == k.Collection
}

func (k *Key) EqualsEntity(src interface{}) bool {
	return k.Equals(NewKey(src))
}

func (k Key) Equals(other *Key) bool {
	if other == nil ||  k.ID == ""  || other.ID == "" {
		return false
	}
	return k.ParentPath == other.ParentPath && k.Collection == other.Collection && k.ID == other.ID
}

func (k *Key) Update(fullPath string) {
	r := regexp.MustCompile(`/documents/([^/]+)/([^/]+)$`)
	if r.MatchString(fullPath) {
		matchs := r.FindAllStringSubmatch(fullPath, -1)
		k.ParentPath = ""
		k.Collection = matchs[0][1]
		k.ID = matchs[0][2]
		return
	}
	r = regexp.MustCompile(`/documents/(.*)/([^/]+)/([^/]+)$`)
	if r.MatchString(fullPath) {
		matchs := r.FindAllStringSubmatch(fullPath, -1)
		k.ParentPath = matchs[0][1]
		k.Collection = matchs[0][2]
		k.ID = matchs[0][3]
		return
	}
	r = regexp.MustCompile(`^(.*)?/([^/]+)/([^/]+)$`)
	if r.MatchString(fullPath) {
		matchs := r.FindAllStringSubmatch(fullPath, -1)
		k.ParentPath = matchs[0][1]
		k.Collection = matchs[0][2]
		k.ID = matchs[0][3]
		return
	}
}

func (k Key) SelfPath() string {
	if k.ID == "" {
		return k.Collection
	}
	return fmt.Sprintf("%s/%s", k.Collection, k.ID)
}

func (k Key) Path() string {
	if k.ParentPath != "" {
		return fmt.Sprintf("%s/%s", k.ParentPath, k.SelfPath())
	}
	return k.SelfPath()
}

func (k Key) URI() string {
	return k.Path()
}

func (k Key) CollectionPath() string {
	if k.ParentPath != "" {
		return fmt.Sprintf("%s/%s", k.ParentPath, k.Collection)
	}
	return k.Collection
}

func (k Key) CreateDocumentRef(client *firestore.Client) *firestore.DocumentRef {
	return client.Doc(k.Path())
}

func (k Key) CreateCollectionRef(client *firestore.Client) *firestore.CollectionRef {
	return client.Collection(k.CollectionPath())
}

func (k Key) CreateGroupCollectionRef(client *firestore.Client) *firestore.CollectionGroupRef {
	return client.CollectionGroup(k.Collection)
}

func (k Key) HasUniqueID() bool {
	return k.ID != ""
}
