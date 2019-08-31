package foon

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"time"
)

type fields struct {
	collection collectionField
	id         *idField
	parent     *parentField
	self       *Key
	createdAt  *createDate
	updaetdAt  *updateDate
}

func getIdField(value reflect.Value) string {
	if value.Type().Kind() == reflect.Ptr {
		return getIdField(reflect.Indirect(value))
	}
	return getIdFieldFromType(value.Type())
}

func getIdFieldFromType(types reflect.Type) string {
	for i := 0 ; i < types.NumField(); i ++ {
		field := types.Field(i)
		if field.Type.Kind() == reflect.Struct {
			if id := getIdFieldFromType(field.Type); id != "" {
				return id
			}
		}
		if field.Tag.Get("foon") == "id" {
			if field.Tag.Get("firestore") != "" {
				return field.Tag.Get("firestore")
			}
			return field.Name
		}
	}
	return ""
}

func newFields(src interface{}) (*fields, error) {
	v := reflect.Indirect(reflect.ValueOf(src)).Type()
	res := &fields{}
	res.collection = newCollectionField(v)
	id, err := newIDField(src)
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
	FieldName string
	ID        string
	field     *reflect.Value
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
	field  *reflect.Value
	parent *Key
}

func NewParentField(src interface{}) *parentField {
	value, _, _ := getField(src, "parent")
	if value == nil {
		return nil
	}
	if key, ok := value.Interface().(*Key); ok {
		return &parentField{value, key}
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
	field, kind, name := getField(src, "id")
	if field != nil {
		if kind == reflect.String {
			return &idField{name, field.String(), field}, nil
		}
		return nil, errors.New("foon id type is must be string")
	}
	return &idField{"", "", nil}, nil
}

func newDateField(src interface{}, foonType string) *dateField {
	value, kind , _ := getField(src, foonType)
	if kind == reflect.Struct && value != nil {
		if _, ok := value.Interface().(time.Time); ok {
			return &dateField{value}
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
		return time.Unix(0, 0)
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
		if f.parent.parent != nil {
			f.parent.parent.Update(ref.Parent.Parent.Path)
		}
		if f.parent.field != nil {
			f.parent.field.Set(reflect.ValueOf(f.parent.parent))
		}
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

func getField(src interface{}, foonName string) (*reflect.Value, reflect.Kind, string) {
	v := reflect.Indirect(reflect.ValueOf(src))
	return findField(v, "foon", foonName)
}

func findField(value reflect.Value, tagName string, tagValue string) (*reflect.Value, reflect.Kind, string) {
	typeName := value.Type()
	fieldNum := typeName.NumField()
	for i := 0; i < fieldNum; i++ {
		field := typeName.Field(i)
		if field.Tag.Get(tagName) == tagValue {
			f := value.Field(i)
			return &f, field.Type.Kind(), field.Name
		}
	}
	for i := 0; i < fieldNum; i++ {
		field := typeName.Field(i)
		if field.Type.Kind() == reflect.Struct {
			nestValue := reflect.Indirect(value.Field(i))
			return findField(nestValue, tagName, tagValue)
		}
	}

	return nil, reflect.Invalid, ""
}

func (i *idField) SetID(id string) error {
	i.ID = id
	if i.field != nil && i.field.CanSet() {
		i.field.SetString(id)
		return nil
	}
	return fmt.Errorf("i.field: %+v canSet: %v", i.field, i.field.CanSet())
}
