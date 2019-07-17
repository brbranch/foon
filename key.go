package foon

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"regexp"
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

func (k Key) Equals(other *Key) bool {
	if k.ID == ""  || other.ID == "" {
		return false
	}
	return k.ParentPath == other.ParentPath && k.Collection == other.Collection && k.ID == other.ID
}

func (k *Key) Update(fullPath string) {
	r := regexp.MustCompile(`/documents/(.*)?/([^/]+)/(.+)$`)
	if r.MatchString(fullPath) {
		matchs := r.FindAllStringSubmatch(fullPath, -1)
		k.ParentPath = matchs[0][1]
		k.Collection = matchs[0][2]
		k.ID = matchs[0][3]
		return
	}
	r = regexp.MustCompile(`^(.*)?/([^/]+)/(.+)$`)
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

func (k Key) HasUniqueID() bool {
	return k.ID != ""
}
