package foon

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

type embedded struct {
	ID string `foon:"id"`
	CreatedAt time.Time `foon:"createdAt"`
	UpdatedAt time.Time `foon:"updatedAt"`
}

type testNormal struct {
	__kind string `foon:"collection,KindTest"`
	ID string `foon:"id"`
	Name string `firestore:"name"`
	CreatedAt time.Time `foon:"createdAt"`
	UpdatedAt time.Time `foon:"updatedAt"`
}

type testEmbedded struct {
	__kind string `foon:"collection,KindEmbedded"`
	embedded
	Name string `firestore:"name"`
}

func TestCollection_ノーマルのテスト(t *testing.T) {
	normal := &testNormal{ID: "test", Name: "name"}
	col , err := newFields(normal)
	assert.NoError(t, err, "failed to create fields")
	assert.Equal(t, "KindTest", col.CollectionName())
}

func TestCollection_Embeddedがされてる場合のテスト(t *testing.T) {
	normal := &testEmbedded{embedded: embedded{ID: "test"}, Name: "name"}
	col , err := newFields(normal)
	assert.NoError(t, err, "failed to create fields")
	assert.Equal(t, "KindEmbedded", col.CollectionName())
}

func TestIdField_ノーマルのテスト(t *testing.T) {
	test := &testNormal{ID: "test"}
	fld , err := newIDField(test)
	assert.NoError(t, err , "Failed to get Fields")

	assert.Equal(t , "test", fld.ID)
	assert.Equal(t , "test", test.ID)
	fld.SetID("bcde")
	assert.Equal(t , "bcde", fld.ID)
	assert.Equal(t , "bcde", test.ID)
}

func TestIdField_Embeddedがされてる場合のテスト(t *testing.T) {
	test := &testEmbedded{embedded: embedded{ID: "test"}, Name: "name"}
	fld , err := newIDField(test)
	assert.NoError(t, err , "Failed to get Fields")

	assert.Equal(t , "test", test.ID, "data failed")
	if !assert.Equal(t , "test", fld.ID, "field is not reflected.") {
		t.Fatal("field is not reflected")
	}

	fld.SetID("bcde")
	print("bbb")
	assert.Equal(t , "bcde", test.ID, "data is not changed.")
	assert.Equal(t , "bcde", fld.ID, "field is not changed.")
}

func TestDateField_正常に変更されるかどうか(t *testing.T) {
	test := &testNormal{ID: "test"}
	create := newCreateField(test)
	update  := newUpdatedField(test)

	assert.True(t, create.has(), "createdAt is not initialized")
	assert.True(t, update.has(), "updatedAt is not initialized")
	assert.True(t, test.CreatedAt.IsZero(), "data is invalid")
	assert.True(t, test.UpdatedAt.IsZero(), "data is invalid")
	assert.True(t, create.get().IsZero(), "creaetdAt is not zero")
	assert.True(t, update.get().IsZero(), "updatedAt is not zero")

	now := time.Now()
	create.UpdateTime(now)
	update.UpdateTime(now)
	assert.Equal(t, create.get(), now, "createdAt is not synced.")
	assert.Equal(t, update.get(), now, "updatedAt is not synced.")
	assert.Equal(t, test.CreatedAt, now, "createdAt(data) is not synced.")
	assert.Equal(t, test.UpdatedAt , now, "updatedAt(data) is not synced.")

	now2 := now.Add(300)
	create.UpdateTime(now2)
	update.UpdateTime(now2)
	assert.Equal(t, create.get(), now, "createdAt is not synced.")
	assert.Equal(t, update.get(), now2, "updatedAt is not synced.")
	assert.Equal(t, test.CreatedAt, now, "createdAt(data) is not synced.")
	assert.Equal(t, test.UpdatedAt , now2, "updatedAt(data) is not synced.")

}

func TestDateField_Embeddedも正常に変更されるかどうか(t *testing.T) {
	test := &testEmbedded{embedded: embedded{ID: "test"}, Name: "name"}
	info , err := newFields(test)

	assert.NoError(t, err , "failed to create fields")

	assert.True(t, info.createdAt.has(), "createdAt is not initialized")
	assert.True(t, info.updaetdAt.has(), "updatedAt is not initialized")
	assert.True(t, test.CreatedAt.IsZero(), "data is invalid")
	assert.True(t, test.UpdatedAt.IsZero(), "data is invalid")
	assert.True(t, info.createdAt.get().IsZero(), "creaetdAt is not zero")
	assert.True(t, info.updaetdAt.get().IsZero(), "updatedAt is not zero")

	now := time.Now()
	info.UpdateTime(now)
	assert.Equal(t, info.createdAt.get(), now, "createdAt is not synced.")
	assert.Equal(t, info.updaetdAt.get(), now, "updatedAt is not synced.")
	assert.Equal(t, test.CreatedAt, now, "createdAt(data) is not synced.")
	assert.Equal(t, test.UpdatedAt , now, "updatedAt(data) is not synced.")

	now2 := now.Add(300)
	info.UpdateTime(now2)
	assert.Equal(t, info.createdAt.get(), now, "createdAt is not synced.")
	assert.Equal(t, info.updaetdAt.get(), now2, "updatedAt is not synced.")
	assert.Equal(t, test.CreatedAt, now, "createdAt(data) is not synced.")
	assert.Equal(t, test.UpdatedAt , now2, "updatedAt(data) is not synced.")

}
