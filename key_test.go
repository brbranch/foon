package foon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type InjectParent struct {
	_kind string `foon:"collection,Parent"`
	ID    string `foon:"id"`
}

type InjectTest struct {
	_kind  string `foon:"collection,TestCollection"`
	ID     string `foon:"id"`
	Parent *Key   `foon:"parent"`
}

func TestNewKeyWithPath(t *testing.T) {
	key := NewKeyWithPath("projects/testProject/databases/(default)/documents/TestCollection/Test001")
	assert.Equal(t, "", key.ParentPath)
	assert.Equal(t, "TestCollection", key.Collection)
	assert.Equal(t, "Test001", key.ID)
}

func TestNewKeyWithPathWithParent(t *testing.T) {
	key := NewKeyWithPath("projects/testProject/databases/(default)/documents/Parent/Parent001/TestCollection/Test001")
	assert.Equal(t, "Parent/Parent001", key.ParentPath)
	assert.Equal(t, "TestCollection", key.Collection)
	assert.Equal(t, "Test001", key.ID)
}

func TestKey_Inject(t *testing.T) {
	key := NewKeyWithPath("projects/testProject/databases/(default)/documents/Parent/Parent001/TestCollection/Test001")
	res := &InjectTest{}
	if err := key.Inject(res); err != nil {
		t.Fatalf("failed to inject (reason: %v)", err)
	}
	assert.Equal(t, "Test001", res.ID)
	if assert.NotNil(t, res.Parent) {
		assert.Equal(t, "Parent001", res.Parent.ID)
		assert.Equal(t, "Parent", res.Parent.Collection)
	}
}

func TestKey_Inject_孫まで(t *testing.T) {
	key := NewKeyWithPath("projects/testProject/databases/(default)/documents/Anc/Anc001/Parent/Parent001/TestCollection/Test001")
	res := &InjectTest{}
	if err := key.Inject(res); err != nil {
		t.Fatalf("failed to inject (reason: %v)", err)
	}
	assert.Equal(t, "Test001", res.ID)
	if assert.NotNil(t, res.Parent) {
		assert.Equal(t, "Parent001", res.Parent.ID)
		assert.Equal(t, "Parent", res.Parent.Collection)
		assert.Equal(t, "Anc/Anc001", res.Parent.ParentPath)

		anc := res.Parent.ParentKey()
		if assert.NotNil(t, anc) {
			assert.Equal(t, "Anc001", anc.ID)
			assert.Equal(t, "Anc", anc.Collection)
		}
	}
}
