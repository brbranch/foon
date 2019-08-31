package foon

import (
		"github.com/stretchr/testify/assert"
	"google.golang.org/appengine/aetest"
	"os"
		"testing"
	"time"
	"cloud.google.com/go/firestore"
)

type TestUser struct {
	__kind    string    `foon:"collection,TestUser"`
	UserID    string    `foon:"id"`
	UserName  string    `firestore:"userName"`
	Age		  int 		`firestore:"age"`
	CreatedAt time.Time `foon:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `foon:"updatedAt" firestore:"updatedAt"`
}

type TestDevice struct {
	__kind     string    `foon:"collection,TestDevice"`
	DeviceID   string    `foon:"id"`
	Parent     *Key      `foon:"parent"`
	DeviceName string    `firestore:"deviceName"`
	CreatedAt  time.Time `foon:"createdAt" firestore:"createdAt"`
	UpdatedAt  time.Time `foon:"updatedAt" firestore:"updatedAt"`
}

type TestMap struct {
	__kind string                `foon:"collection,TestMap"`
	ID     string                `foon:"id"`
	Map    map[string]TestStruct `foon:"testMap"`
}

type TestStruct struct {
	A int
	B string
	C []string
}

func Test_Insertは重複を許可しない(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserID:   "abcde",
		UserName: "kazuki",
	}

	if err := store.Delete(user); err != nil {
		t.Fatalf("failed to delete (reason: %v)", err)
	}

	if err := store.Insert(user); err != nil {
		t.Fatalf("failed to insert (reason: %v)", err)
	}

	if err := store.Insert(user); err == nil {
		t.Fatalf("expected is error but no error is occurred.")
	}
}

func Test_InsertMultiは重複を許可しない(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserID:   "abcde",
		UserName: "kazuki",
	}
	user2 := &TestUser{
		UserID:   "abcdef",
		UserName: "kazuki",
	}
	user3 := &TestUser{
		UserID:   "abcde",
		UserName: "kazuki",
	}
	if err := store.Delete(user); err != nil {
		t.Fatalf("failed to delete (reason: %v)", err)
	}
	if err := store.Delete(user2); err != nil {
		t.Fatalf("failed to delete (reason: %v)", err)
	}

	if err := store.InsertMulti(&[]interface{}{user, user2}); err != nil {
		t.Fatalf("failed to insert (reason: %v)", err)
	}
	if err := store.Delete(user); err != nil {
		t.Fatalf("failed to delete (reason: %v)", err)
	}
	if err := store.Delete(user2); err != nil {
		t.Fatalf("failed to delete (reason: %v)", err)
	}

	if err := store.InsertMulti(&[]interface{}{user, user2, user3}); err == nil {
		t.Fatalf("expected is error but no error is occurred.")
	}
}

func Test_通常のEntityをPutした後Getする(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	user2 := &TestUser{
		UserID: user.UserID,
	}

	if err := store.Get(user2); err != nil {
		t.Fatalf("failed to get user (reason: %v)", err)
	}

	assert.Equal(t, user.UserID, user2.UserID)
	assert.Equal(t, user.UserName, user2.UserName)
	assert.Equal(t, user.CreatedAt.Unix(), user2.CreatedAt.Unix())
	assert.Equal(t, user.UpdatedAt.Unix(), user2.UpdatedAt.Unix())

	user3 := &TestUser{
		UserID: user.UserID,
	}

	if err := store.GetWithoutCache(user3); err != nil {
		t.Fatalf("failed to get user (reason: %v)", err)
	}

	assert.Equal(t, user.UserID, user3.UserID)
	assert.Equal(t, user.UserName, user3.UserName)
	assert.Equal(t, user.CreatedAt.Unix(), user3.CreatedAt.Unix())
	assert.Equal(t, user.UpdatedAt.Unix(), user3.UpdatedAt.Unix())

	user.UserName = "kazuki2"

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to update user (reason: %v)", err)
	}

	user4 := &TestUser{
		UserID: user.UserID,
	}

	if err := store.GetWithoutCache(user4); err != nil {
		t.Fatalf("failed to get user (reason: %v)", err)
	}

	assert.Equal(t, user.UserID, user4.UserID)
	assert.Equal(t, user.UserName, user4.UserName)
	assert.Equal(t, user.CreatedAt.Unix(), user4.CreatedAt.Unix())
	assert.Equal(t, user.UpdatedAt.Unix(), user4.UpdatedAt.Unix())

}

func Test_Map形式のも入れられる(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	testMap := &TestMap{
		ID: "abcde",
		Map: map[string]TestStruct{
			"B": {
				A: 1,
				B: "acd",
				C: []string{"a", "b", "c"},
			},
		},
	}

	if err := store.Put(testMap); err != nil {
		t.Fatalf("failed to put store : %v", err)
	}

	maps := &TestMap{ID: "abcde"}
	if err := store.Get(maps); err != nil {
		t.Fatalf("failed to get store : %v", err)
	}

	assert.Equal(t, "abcde", maps.ID)
	assert.Equal(t, 1, maps.Map["B"].A)
	assert.Equal(t, "acd", maps.Map["B"].B)
	assert.Equal(t, []string{"a", "b", "c"}, maps.Map["B"].C)

}

func Test_ChildのEntityをPutした後Getする(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	device1 := &TestDevice{
		Parent:     NewKey(user),
		DeviceName: "Device001",
	}

	if err := store.Put(device1); err != nil {
		t.Fatalf("failed to put device (reason: %v)", err)
	}

	assert.NotEmpty(t, user.UserID)
	assert.NotEmpty(t, device1.DeviceID)

	device2 := &TestDevice{
		Parent:   NewKey(user),
		DeviceID: device1.DeviceID,
	}

	if err := store.Get(device2); err != nil {
		t.Fatalf("failed to get device (reason: %v)", err)
	}

	assert.Equal(t, device1.DeviceID, device2.DeviceID)
	assert.Equal(t, device1.DeviceName, device2.DeviceName)
	assert.Equal(t, device1.UpdatedAt.Unix(), device2.UpdatedAt.Unix())
	assert.Equal(t, device1.CreatedAt.Unix(), device2.CreatedAt.Unix())
	assert.Equal(t, "", device2.Parent.ParentPath)
	assert.Equal(t, "TestUser", device2.Parent.Collection)
	assert.Equal(t, user.UserID, device2.Parent.ID)
	assert.Equal(t, device1.DeviceID, device2.DeviceID)

	device3 := &TestDevice{
		Parent:   NewKey(user),
		DeviceID: device1.DeviceID,
	}

	if err := store.GetWithoutCache(device3); err != nil {
		t.Fatalf("failed to get device without cache (reason: %v)", err)
	}

	assert.Equal(t, device1.DeviceID, device3.DeviceID)
	assert.Equal(t, device1.DeviceName, device3.DeviceName)
	assert.Equal(t, device1.UpdatedAt.Unix(), device3.UpdatedAt.Unix())
	assert.Equal(t, device1.CreatedAt.Unix(), device3.CreatedAt.Unix())
	assert.Equal(t, "", device3.Parent.ParentPath)
	assert.Equal(t, "TestUser", device3.Parent.Collection)
	assert.Equal(t, user.UserID, device3.Parent.ID)
	assert.Equal(t, device1.DeviceID, device3.DeviceID)

	device1.DeviceName = "ChangeDeviceName"
	if err := store.Put(device1); err != nil {
		t.Fatalf("failed to save data (reason:%v)", err)
	}

	device4 := &TestDevice{
		Parent:   NewKey(user),
		DeviceID: device1.DeviceID,
	}

	if err := store.GetWithoutCache(device4); err != nil {
		t.Fatalf("failed to get device without cache (reason: %v)", err)
	}

	assert.Equal(t, device1.DeviceID, device4.DeviceID)
	assert.Equal(t, "ChangeDeviceName", device4.DeviceName)
	assert.Equal(t, device1.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
	assert.Equal(t, device1.CreatedAt.Unix(), device4.CreatedAt.Unix())
	assert.Equal(t, "", device4.Parent.ParentPath)
	assert.Equal(t, "TestUser", device4.Parent.Collection)
	assert.Equal(t, user.UserID, device4.Parent.ID)
	assert.Equal(t, device1.DeviceID, device4.DeviceID)

}

func Test_PutMulti(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	key := NewKey(user)
	devices := []*TestDevice{
		{Parent: key, DeviceName: "Device001"},
		{Parent: key, DeviceName: "Device002"},
		{Parent: key, DeviceName: "Device003"},
	}

	if err := store.PutMulti(devices); err != nil {
		t.Fatalf("failed to putMulti (reason: %v)", err)
	}

	for _, device := range devices {
		device4 := &TestDevice{
			DeviceID: device.DeviceID,
			Parent:   key,
		}

		if err := store.Get(device4); err != nil {
			t.Fatalf("failed to get data (reason:%v)", err)
		}

		assert.False(t, device.UpdatedAt.IsZero())
		assert.False(t, device.CreatedAt.IsZero())

		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

}

func Test_GetMulti(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserID:   "kazuki.oda",
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	key := NewKey(user)
	deleteCache := &TestDevice{Parent: key, DeviceID: "device04", DeviceName: "Device004"}
	devices := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01", DeviceName: "Device001"},
		&TestDevice{Parent: key, DeviceID: "device02", DeviceName: "Device002"},
		&TestDevice{Parent: key, DeviceID: "device03", DeviceName: "Device003"},
		deleteCache,
	}

	if err := store.PutMulti(devices); err != nil {
		t.Fatalf("failed to putMulti (reason: %v)", err)
	}

	store.cache.Delete(NewKey(deleteCache))

	devCache := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01"},
		&TestDevice{Parent: key, DeviceID: "device02"},
		&TestDevice{Parent: key, DeviceID: "device04"},
	}

	if err := store.GetMulti(&devCache); err != nil {
		t.Fatalf("failed to get multi (reason: %v)", err)
	}

	results := []int{0, 1, 3}

	for idx, device4 := range devCache {

		device := devices[results[idx]]
		assert.False(t, device.UpdatedAt.IsZero())
		assert.False(t, device.CreatedAt.IsZero())

		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

	devCache2 := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01"},
		&TestDevice{Parent: key, DeviceID: "device02"},
		&TestDevice{Parent: key, DeviceID: "device04"},
	}

	if err := store.GetMulti(&devCache2); err != nil {
		t.Fatalf("failed to get multi (reason: %v)", err)
	}

	for idx, device4 := range devCache2 {

		device := devices[results[idx]]
		assert.False(t, device.UpdatedAt.IsZero())
		assert.False(t, device.CreatedAt.IsZero())

		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

	devs := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01"},
		&TestDevice{Parent: key, DeviceID: "device02"},
		&TestDevice{Parent: key, DeviceID: "device03"},
		&TestDevice{Parent: key, DeviceID: "device04"},
	}

	if err := store.GetMultiWithoutCache(&devs); err != nil {
		t.Fatalf("failed to get multi (reason: %v)", err)
	}

	results = []int{0, 1, 2, 3}

	for idx, device4 := range devs {

		device := devices[results[idx]]
		assert.False(t, device.UpdatedAt.IsZero())
		assert.False(t, device.CreatedAt.IsZero())

		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

}

func Test_Condition_Where(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	users := []*TestUser {
		{UserID: "user001", UserName: "username001", Age: 1},
		{UserID: "user002", UserName: "username002", Age: 2},
		{UserID: "user003", UserName: "username003", Age: 3},
		{UserID: "user004", UserName: "username004", Age: 4},
	}

	if err := store.PutMulti(&users); err != nil {
		t.Fatalf("failed to put test data.")
	}

	key := NewKey(&TestUser{})
	results := []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().Where("userName" , "==", "username001")); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,1, len(results))
	assert.Equal(t, "user001", results[0].UserID)
	assert.Equal(t, "username001", results[0].UserName)
	assert.Equal(t, 1, results[0].Age)

	// cacheが使われる
	results = []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().Where("userName" , "==", "username001")); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,1, len(results))
	assert.Equal(t, "user001", results[0].UserID)
	assert.Equal(t, "username001", results[0].UserName)
	assert.Equal(t, 1, results[0].Age)

	// cacheは使われない
	results = []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().Where("age" , ">", 2)); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,2, len(results))
	assert.Equal(t, "user003", results[0].UserID)
	assert.Equal(t, "username003", results[0].UserName)
	assert.Equal(t, 3, results[0].Age)

}

func Test_Condition_Limit(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	users := []*TestUser {
		{UserID: "user001", UserName: "username001", Age: -1},
		{UserID: "user002", UserName: "username002", Age: 2},
		{UserID: "user003", UserName: "username003", Age: 3},
		{UserID: "user004", UserName: "username004", Age: 499},
	}

	if err := store.PutMulti(&users); err != nil {
		t.Fatalf("failed to put test data.")
	}

	key := NewKey(&TestUser{})
	results := []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().Limit(1).OrderBy("age", firestore.Asc)); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,1, len(results))
	assert.Equal(t, "user001", results[0].UserID)
	assert.Equal(t, "username001", results[0].UserName)
	assert.Equal(t, -1, results[0].Age)

	// cacheは使われる
	results = []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().OrderBy("age", firestore.Asc).Limit(1)); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,1, len(results))
	assert.Equal(t, "user001", results[0].UserID)
	assert.Equal(t, "username001", results[0].UserName)
	assert.Equal(t, -1, results[0].Age)

	// cacheは使われない
	results = []*TestUser{}

	if err := store.GetByQuery(key, &results, NewConditions().OrderBy("age", firestore.Desc).Limit(1)); err != nil {
		t.Fatalf("failed to get data.")
	}

	assert.Equal(t,1, len(results))
	assert.Equal(t, "user004", results[0].UserID)
	assert.Equal(t, "username004", results[0].UserName)
	assert.Equal(t, 499, results[0].Age)
}

func Test_GetAll(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserID:   "kazuki.oda",
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	key := NewKey(user)
	devices := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01", DeviceName: "Device001"},
		&TestDevice{Parent: key, DeviceID: "device02", DeviceName: "Device002"},
		&TestDevice{Parent: key, DeviceID: "device03", DeviceName: "Device003"},
		&TestDevice{Parent: key, DeviceID: "device04", DeviceName: "Device004"},
	}

	if err := store.PutMulti(devices); err != nil {
		t.Fatalf("failed to putMulti (reason: %v)", err)
	}

	devKey := NewKey(&TestDevice{Parent: key})
	dev := []*TestDevice{}

	if err := store.GetAll(devKey, &dev); err != nil {
		t.Fatalf("failed to get condition (reason:%v)", err)
	}

	assert.Equal(t, 4, len(dev))
	for idx, device := range devices {
		device4 := dev[idx]
		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

	dev2 := []*TestDevice{}
	if err := store.GetAll(devKey, &dev2); err != nil {
		t.Fatalf("failed to get condition (reason:%v)", err)
	}

	assert.Equal(t, 4, len(dev2))
	for idx, device := range devices {
		device4 := dev2[idx]
		assert.Equal(t, device.DeviceID, device4.DeviceID)
		assert.Equal(t, device.DeviceName, device4.DeviceName)
		assert.Equal(t, device.UpdatedAt.Unix(), device4.UpdatedAt.Unix())
		assert.Equal(t, device.CreatedAt.Unix(), device4.CreatedAt.Unix())
		assert.Equal(t, "", device4.Parent.ParentPath)
		assert.Equal(t, "TestUser", device4.Parent.Collection)
		assert.Equal(t, user.UserID, device4.Parent.ID)
		assert.Equal(t, device.DeviceID, device4.DeviceID)
	}

}

func TestFoon_GetGroupByQuery(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	user := &TestUser{
		UserID:   "kazuki.oda",
		UserName: "kazuki",
	}

	if err := store.Put(user); err != nil {
		t.Fatalf("failed to put user (reason: %v)", err)
	}

	key := NewKey(user)
	devices := []*TestDevice{
		&TestDevice{Parent: key, DeviceID: "device01", DeviceName: "Device001"},
		&TestDevice{Parent: key, DeviceID: "device02", DeviceName: "Device002"},
		&TestDevice{Parent: key, DeviceID: "device03", DeviceName: "Device003"},
		&TestDevice{Parent: key, DeviceID: "device04", DeviceName: "Device008"},
	}

	if err := store.PutMulti(devices); err != nil {
		t.Fatalf("failed to putMulti (reason: %v)", err)
	}

	dev := []*TestDevice{}

	if err := store.GetGroupByQuery(&dev, NewConditions().Where("deviceName" , "==", "Device008")); err != nil {
		t.Fatalf("failed to get group (reason:%v)", err)
	}

	assert.Equal(t, 1, len(dev))
	assert.Equal(t, "Device008", dev[0].DeviceName)

}

func TestFoon_GetGroupByQuery_複数(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8915")
	store, err := NewStoreWithProjectID(ctx, "everychart-dev")

	if err != nil {
		t.Fatalf("failed to create Foon Client (reason: %v)", err)
	}

	users := []*TestUser {
		{
			UserID:   "user001",
			UserName: "kazuki",
		},
		{
			UserID:   "user002",
			UserName: "kazuki2",
		},
	}

	for _ , user := range users {
		key := NewKey(user)
		devices := []*TestDevice{
			&TestDevice{Parent: key, DeviceID: "Device01", DeviceName: "Mevice001"},
			&TestDevice{Parent: key, DeviceID: "Device02", DeviceName: "Mevice002"},
			&TestDevice{Parent: key, DeviceID: "Device03", DeviceName: "Mevice003"},
			&TestDevice{Parent: key, DeviceID: "Device04", DeviceName: "Mevice004"},
			&TestDevice{Parent: key, DeviceID: "Device05", DeviceName: "Mevice001"},
			&TestDevice{Parent: key, DeviceID: "Device06", DeviceName: "Mevice002"},
			&TestDevice{Parent: key, DeviceID: "Device07", DeviceName: "Mevice003"},
			&TestDevice{Parent: key, DeviceID: "Device08", DeviceName: "Mevice004"},
		}
		for _ , device := range devices {
			if err := store.Delete(device); err != nil {
				t.Fatalf("failed to delete devices (reason: %v)", err)
			}
		}
		if err := store.Delete(user); err != nil {
			t.Fatalf("failed to delete user (reason: %v)", err)
		}
	}

	{
		user := &TestUser{
			UserID:   "user001",
			UserName: "kazuki",
		}

		if err := store.Put(user); err != nil {
			t.Fatalf("failed to put user (reason: %v)", err)
		}

		key := NewKey(user)
		devices := []*TestDevice{
			&TestDevice{Parent: key, DeviceID: "Device01", DeviceName: "Mevice001"},
			&TestDevice{Parent: key, DeviceID: "Device02", DeviceName: "Mevice002"},
			&TestDevice{Parent: key, DeviceID: "Device03", DeviceName: "Mevice003"},
			&TestDevice{Parent: key, DeviceID: "Device04", DeviceName: "Mevice004"},
		}

		if err := store.PutMulti(devices); err != nil {
			t.Fatalf("failed to putMulti (reason: %v)", err)
		}
	}

	{

		dev := []*TestDevice{}

		if err := store.GetGroupByQuery(&dev, NewConditions().Where("deviceName", "==", "Mevice004")); err != nil {
			t.Fatalf("failed to get group (reason:%v)", err)
		}

		assert.Equal(t, 1, len(dev))
		assert.Equal(t, "Device04", dev[0].DeviceID)
		assert.Equal(t, "Mevice004", dev[0].DeviceName)
	}

	{
		user := &TestUser{
			UserID:   "user002",
			UserName: "kazuki2",
		}

		if err := store.Put(user); err != nil {
			t.Fatalf("failed to put user (reason: %v)", err)
		}

		key := NewKey(user)
		devices := []*TestDevice{
			&TestDevice{Parent: key, DeviceID: "Device05", DeviceName: "Mevice001"},
			&TestDevice{Parent: key, DeviceID: "Device06", DeviceName: "Mevice002"},
			&TestDevice{Parent: key, DeviceID: "Device07", DeviceName: "Mevice003"},
			&TestDevice{Parent: key, DeviceID: "Device08", DeviceName: "Mevice004"},
		}

		if err := store.PutMulti(devices); err != nil {
			t.Fatalf("failed to putMulti (reason: %v)", err)
		}
	}

	{

		dev := []*TestDevice{}

		if err := store.GetGroupByQuery(&dev, NewConditions().Where("deviceName", "==", "Mevice004")); err != nil {
			t.Fatalf("failed to get group (reason:%v)", err)
		}

		if assert.Equal(t, 2, len(dev)) {
			assert.Equal(t, "Device04", dev[0].DeviceID)
			assert.Equal(t, "Mevice004", dev[0].DeviceName)
			assert.Equal(t, "user001", dev[0].Parent.ID)
			assert.Equal(t, "Device08", dev[1].DeviceID)
			assert.Equal(t, "Mevice004", dev[1].DeviceName)
			assert.Equal(t, "user002", dev[1].Parent.ID)
		}
	}

	{

		dev := []*TestDevice{}
		user := &TestUser{
			UserID:   "user002",
			UserName: "kazuki2",
		}

		if err := store.GetByQuery(NewKey(&TestDevice{Parent:NewKey(user)}), &dev, NewConditions().Where("deviceName", "==", "Mevice004")); err != nil {
			t.Fatalf("failed to get group (reason:%v)", err)
		}

		if assert.Equal(t, 1, len(dev)) {
			assert.Equal(t, "Device08", dev[0].DeviceID)
			assert.Equal(t, "Mevice004", dev[0].DeviceName)
			assert.Equal(t, "user002", dev[0].Parent.ID)
		}
	}

}
