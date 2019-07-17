package foon

import (
	"testing"
			"github.com/stretchr/testify/assert"
	"context"
	"cloud.google.com/go/firestore"
)

type TestColl struct {
	__kind string `foon:"collection,TestColl"`
	ID string `foon:"id" firestore:"id"`
}

type TestFoon struct {
	__kind string `foon:"collection,InnerTestColl"`
	Parent *Key `foon:"parent" firestore:"-"`
	ID string `foon:"id" firestore:"id"`
}

type TestChild struct {
	__kind string `foon:"collection,InnerTestChild"`
	Parent *Key `foon:"parent" firestore:"-"`
	ID string `foon:"id" firestore:"id"`
}

func TestKeyのPath作成(t *testing.T) {
	parent := &TestColl{ID: "parent001"}
	key1 := NewKey(parent)
	assert.Equal(t, "TestColl/parent001", key1.Path())
	assert.Equal(t, "TestColl/parent001", key1.SelfPath())

	child := &TestFoon{Parent: key1, ID: "child002"}
	key2 := NewKey(child)

	assert.Equal(t, "TestColl/parent001/InnerTestColl/child002", key2.Path())
	assert.Equal(t, "InnerTestColl/child002", key2.SelfPath())

	child2 := &TestChild{Parent: key2, ID: "child003"}
	key3 := NewKey(child2)
	assert.Equal(t, "TestColl/parent001/InnerTestColl/child002/InnerTestChild/child003", key3.Path())
	assert.Equal(t, "InnerTestChild/child003", key3.SelfPath())
}

func Test_PathからKeyを生成(t *testing.T) {
	key1 := NewKeyWithPath("projects/projectID/databases/(default)/documents/TestColl/parent001/InnerTestColl/parent002")
	assert.Equal(t, "TestColl/parent001", key1.ParentPath)
	assert.Equal(t, "InnerTestColl", key1.Collection)
	assert.Equal(t, "parent002", key1.ID)

	key1.Update("Parent/parent999/TestColl/002")
	assert.Equal(t, "Parent/parent999", key1.ParentPath)
	assert.Equal(t, "TestColl", key1.Collection)
	assert.Equal(t, "002", key1.ID)

	key1.Update("/TestColl2/003")
	assert.Equal(t, "", key1.ParentPath)
	assert.Equal(t, "TestColl2", key1.Collection)
	assert.Equal(t, "003", key1.ID)
}

func Testコレクションがネストされる(t *testing.T) {

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "projectID")

	if err != nil {
		t.Fatalf("failed to create firestore (reason: %+v)", err)
	}

	parent := &TestColl{ID: "parent001"}
	key := NewKey(parent)
	foon := &TestFoon{
		Parent: key,
		ID: "parent002",
	}

	childKey := NewKey(foon)
	ref := childKey.CreateDocumentRef(client)

	assert.Equal(t, "projects/projectID/databases/(default)/documents/TestColl/parent001/InnerTestColl/parent002", ref.Path)
}

