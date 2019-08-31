package foon

import (
	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"google.golang.org/appengine/aetest"
	"os"
	"fmt"
)

type CursorTest struct {
	__kind string    `foon:"collection,CursorTest"`
	ID     string    `foon:"id" firestore:"id"`
	Name   string    `firestore:"name"`
	Num    int       `firestore:"num"`
	Time   time.Time `firestore:"time"`
}

func TestCursor_カーソルが正常に働く(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	// 事前準備
	datas := []*CursorTest{}
	now := time.Now()
	for i := 0; i < 12; i++ {
		datas = append(datas, &CursorTest{
			ID: fmt.Sprintf("cursor%03d", i),
			Name: fmt.Sprintf("name%d", i),
			Num: (100 - i) / 2,
			Time: time.Unix(now.Unix() + int64(i) , 0),
		})
	}
	if err := store.PutMulti(&datas); err != nil {
		t.Errorf("failed to put data (reason: %v)", err)
	}
	checkData := []*CursorTest{}
	if err := store.GetAll(NewKey(&CursorTest{}), &checkData); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}

	assert.Equal(t, 12, len(checkData))

	// テスト開始
	condition := NewConditions().OrderBy("num", firestore.Asc).OrderBy("id", firestore.Desc).Limit(5)
	newDatas := []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, condition); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}

	assert.Equal(t, 5, len(newDatas))
	assert.Equal(t, "cursor011", newDatas[0].ID)
	assert.Equal(t, "cursor010", newDatas[1].ID)
	assert.Equal(t, "cursor009", newDatas[2].ID)
	assert.Equal(t, "cursor008", newDatas[3].ID)
	assert.Equal(t, "cursor007", newDatas[4].ID)
	cursor := store.LastCursor()
	fmt.Printf("cursor: %s\n", cursor)
	assert.NotEqual(t, "", cursor)

	nextCursor, err := NewCursor(cursor)
	if err != nil {
		t.Fatalf("failed to create next cursor (reason: %v)", err)
	}

	fmt.Println(nextCursor.planeCursor())

	assert.Equal(t, nextCursor.Path, "CursorTest/cursor007")
	assert.Equal(t, nextCursor.Orders, []CursorOrder{{"num", firestore.Asc}, {"id", firestore.Desc}})

	newCond := NewConditions().StartAfter(nextCursor).Limit(5)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, newCond); err != nil {
		t.Fatalf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 5, len(newDatas))
	assert.Equal(t, "cursor006", newDatas[0].ID)
	assert.Equal(t, "cursor005", newDatas[1].ID)
	assert.Equal(t, "cursor004", newDatas[2].ID)
	assert.Equal(t, "cursor003", newDatas[3].ID)
	assert.Equal(t, "cursor002", newDatas[4].ID)
	cursor = store.LastCursor()
	fmt.Printf("cursor: %s\n", cursor)
	assert.NotEqual(t, "", cursor)

	nextCursor, err = NewCursor(cursor)
	if err != nil {
		t.Fatalf("failed to create next cursor (reason: %v)", err)
	}
	assert.Equal(t, nextCursor.ID, "id")
	assert.Equal(t, nextCursor.Path, "CursorTest/cursor002")
	assert.Equal(t, nextCursor.Orders, []CursorOrder{{"num", firestore.Asc}, {"id", firestore.Desc}})
	fmt.Println(nextCursor.planeCursor())

	nextCondition := NewConditions().StartAfter(nextCursor).Limit(5)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, nextCondition); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 2, len(newDatas))
	assert.Equal(t, "cursor001", newDatas[0].ID)
	assert.Equal(t, "cursor000", newDatas[1].ID)

	cursor = store.LastCursor()
	assert.Equal(t, "", cursor)

	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, newCond); err != nil {
		t.Fatalf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 5, len(newDatas))
	assert.Equal(t, "cursor006", newDatas[0].ID)
	assert.Equal(t, "cursor005", newDatas[1].ID)
	assert.Equal(t, "cursor004", newDatas[2].ID)
	assert.Equal(t, "cursor003", newDatas[3].ID)
	assert.Equal(t, "cursor002", newDatas[4].ID)
	cursor = store.LastCursor()
	fmt.Printf("cursor: %s\n", cursor)
	assert.NotEqual(t, "", cursor)


	timeCond := NewConditions().OrderBy("time", firestore.Asc).Limit(5)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, timeCond); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 5, len(newDatas))
	assert.Equal(t, "cursor000", newDatas[0].ID)
	assert.Equal(t, "cursor001", newDatas[1].ID)
	assert.Equal(t, "cursor002", newDatas[2].ID)
	assert.Equal(t, "cursor003", newDatas[3].ID)
	assert.Equal(t, "cursor004", newDatas[4].ID)

	timeCond = NewConditions().OrderBy("time", firestore.Asc).Limit(5)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, timeCond); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 5, len(newDatas))
	assert.Equal(t, "cursor000", newDatas[0].ID)
	assert.Equal(t, "cursor001", newDatas[1].ID)
	assert.Equal(t, "cursor002", newDatas[2].ID)
	assert.Equal(t, "cursor003", newDatas[3].ID)
	assert.Equal(t, "cursor004", newDatas[4].ID)

	cursor = store.LastCursor()
	assert.NotEqual(t, "", cursor)
	nextCursor, err = NewCursor(cursor)
	if err != nil {
		t.Fatalf("failed to create next cursor (reason: %v)", err)
	}

	timeCond = NewConditions().StartAfter(nextCursor).Limit(7)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, timeCond); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 7, len(newDatas))
	assert.Equal(t, "cursor005", newDatas[0].ID)
	assert.Equal(t, "cursor006", newDatas[1].ID)
	assert.Equal(t, "cursor007", newDatas[2].ID)
	assert.Equal(t, "cursor008", newDatas[3].ID)
	assert.Equal(t, "cursor009", newDatas[4].ID)
	assert.Equal(t, "cursor010", newDatas[5].ID)
	assert.Equal(t, "cursor011", newDatas[6].ID)

	cursor = store.LastCursor()
	assert.NotEqual(t, "", cursor)
	nextCursor, err = NewCursor(cursor)
	if err != nil {
		t.Fatalf("failed to create next cursor (reason: %v)", err)
	}
	timeCond = NewConditions().StartAfter(nextCursor).Limit(7)
	newDatas = []*CursorTest{}
	if err := store.GetByQuery(NewKey(&CursorTest{}), &newDatas, timeCond); err != nil {
		t.Errorf("failed to get data (reason: %v)", err)
	}
	assert.Equal(t, 0, len(newDatas))
	cursor = store.LastCursor()
	assert.Equal(t, "", cursor)
}
