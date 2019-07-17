package foon

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"time"
	"fmt"
	"cloud.google.com/go/firestore"
)

type fields struct {
	collection collectionField
	id *idField
	parent *parentField
	self *Key
	createdAt *createDate
	updaetdAt *updateDate
}

func newFields(src interface{}) (*fields, error) {
	v := reflect.Indirect(reflect.ValueOf(src)).Type()
	res := &fields{}
	res.collection = newCollectionField(v)
	id , err := newIDField(src)
	if err != nil {
		return nil, err
	}
	res.id = id
	res.parent = NewParentField(src)
	res.self = nil
	res.createdAt = newCreateField(src)
	res.updaetdAt = newUpdatedField(src)
	return res, nil
}

func (f fields) CollectionName() string {
	return f.collection.Name
}

func (f fields) UpdateTime(now time.Time) {
	f.createdAt.UpdateTime(now)
	f.updaetdAt.UpdateTime(now)
}

type collectionField struct {
	Name string
}

type idField struct {
	ID    string
	field *reflect.Value
}

type dateField struct {
	field *reflect.Value
}

type createDate struct {
	dateField
}

type updateDate struct {
	dateField
}

type parentField struct {
	field *reflect.Value
	parent *Key
}

func NewParentField(src interface{}) *parentField {
	value, _ := getField(src, "parent")
	if value == nil {
		return nil
	}
	if key, ok := value.Interface().(*Key); ok {
		return  &parentField{value, key}
	}

	return nil
}

func (f fields) HasUniqueID() bool {
	return f.id != nil && f.id.ID != ""
}

func (f *fields) SetID(newID string) {
	if f.id != nil {
		f.id.SetID(newID)
	}
	if f.self != nil {
		f.self.ID = newID
	}
}

func (f *fields) UpdateField(newID string, updateTime time.Time) {
	f.SetID(newID)
	f.UpdateTime(updateTime)
}

func newCollectionField(t reflect.Type) collectionField {
	for i := 0; i < t.NumField(); i++ {
		val := t.Field(i)
		tag := val.Tag.Get("foon")
		if len(tag) > 11 {
			if tag[:10] == "collection" {
				return collectionField{tag[11:]}
			}
		}
	}
	name := t.String()
	name = name[strings.Index(name, ".")+1:]
	return collectionField{name}
}

func newIDField(src interface{}) (*idField, error) {
	field, kind := getField(src, "id")
	if field != nil {
		if kind == reflect.String {
			return &idField{field.String(), field}, nil
		}
		return nil, errors.New("foon id type is must be string")
	}
	return &idField{"", nil}, nil
}

func newDateField(src interface{}, foonType string) *dateField {
	field, kind := getField(src, foonType)
	if kind == reflect.Struct && field != nil{
		if _ , ok := field.Interface().(time.Time); ok {
			return &dateField{field}
		}
	}
	return &dateField{nil}
}

func newCreateField(src interface{}) *createDate {
	return &createDate{*newDateField(src, "createdAt")}
}

func newUpdatedField(src interface{}) *updateDate {
	return &updateDate{*newDateField(src, "updatedAt")}
}

func (f *dateField) has() bool {
	return f.field != nil
}

func (f *dateField) get() time.Time {
	if f.field == nil {
		return time.Unix(0,0)
	}
	return f.field.Interface().(time.Time)
}

func (f *dateField) set(now time.Time) {
	if f.field == nil {
		return
	}
	f.field.Set(reflect.ValueOf(now))
}

func (f *fields) updateKey(ref *firestore.DocumentRef) {
	if f.parent != nil {
		f.parent.parent.Update(ref.Parent.Parent.Path)
		f.parent.field.Set(reflect.ValueOf(f.parent.parent))
	}
}

func (f *createDate) UpdateTime(now time.Time) {
	if f.has() && f.get().IsZero() {
		f.set(now)
	}
}

func (f *updateDate) UpdateTime(now time.Time) {
	f.set(now)
}

func getField(src interface{}, foonName string) (*reflect.Value, reflect.Kind) {
	v := reflect.Indirect(reflect.ValueOf(src))
	return findField(v, "foon", foonName)
}

func findField(value reflect.Value, tagName string, tagValue string) (*reflect.Value, reflect.Kind) {
	typeName := value.Type()
	fieldNum := typeName.NumField()
	for i := 0; i < fieldNum; i++ {
		field := typeName.Field(i)
		if field.Tag.Get(tagName) == tagValue {
			f := value.Field(i)
			return &f, field.Type.Kind()
		}
	}
	for i := 0; i < fieldNum; i++ {
		field := typeName.Field(i)
		if field.Type.Kind() == reflect.Struct {
			nestValue := reflect.Indirect(value.Field(i))
			return findField(nestValue, tagName, tagValue)
		}
	}

	return nil, reflect.Invalid
}

func (i *idField) SetID(id string) error {
	i.ID = id
	if i.field != nil && i.field.CanSet() {
		i.field.SetString(id)
		return nil
	}
	return fmt.Errorf("i.field: %+v canSet: %v", i.field, i.field.CanSet())
}