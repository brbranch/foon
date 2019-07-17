# foon
Foon is a wrapper library for Firestore Native API and which provides an autocaching interface to the Google App Engine. Foon can use like [mjibson/goon](https://github.com/mjibson/goon).

## Usage
### Struct
Below is an example struct which are stored to Firestore Document (in "Users" Collection).

```go
// Collection
type User struct {
	__kind    string    `foon:"collection,Users"` // Specifies Collection Name
	ID        string    `foon:"id"` // When the document was put into Firestore, this field stores Document ID automaticary.
	Name      string    `firestore:"name"`
	CreatedAt time.Time `foon:"createdAt" firestore:"createdAt"` // When the document was put into Firestore, this field stores stored datetime.
	UpdatedAt time.Time `foon:"updatedAt" firestore:"updatedAt"` // When the document was put or updated, this field modified automaticary.
}

// Sub Collection
type Device struct {
	__kind     string    `foon:"collection,Devices"`
	ID         string    `foon:"id"`
	Parent     *foon.Key `foon:"parent"`
	DeviceName string    `firestore:"deviceName"`
}
```

### Get Document

Following are examples to get document from "User" collection which are stored in Firestore.

```go
user := &User{ ID: "user001" }
f := foon.Must(appEngineContext)
if err := f.Get(user); err != nil {
    log.Warningf(appEngineContext, "failed to get user.")
}
```

e.g If Devices are the Sub Collection from User Document, followings are usage to get some Device document.

```go
user := &User{ ID: "user001" }
key := foon.NewKey(user)

device := &Device{ID: "device001", Parent: key }
f := foon.Must(appEngineContext)
f.Get(device)
```

### Put Document
Following are examples.

```go
f := foon.Must(appEngineContext)
user := &User{ Name: "username01" }

// if you don't specify id, foon stores random id automaticary.
if err := f.Insert(user); err == nil {
    // store sub collection
    device := &Device{ ID: "device01", Parent: foon.NewKey(user) }
    f.Put(device)
}

```