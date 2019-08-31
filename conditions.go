package foon

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"bytes"
		"sort"
	"crypto/md5"
	)

type Query interface {
	Queryer(query firestore.Query) firestore.Query
	Hash() string
	Order() int
}

type CursorQuery interface {
	Queryer(query firestore.Query, cursor *Cursor, f *Foon) (firestore.Query, error)
	Hash(cursor *Cursor) string
}

type Queries []Query

type Conditions struct {
	limit int
	cursor *Cursor
	cursorQuery CursorQuery
	group   string
	Queries Queries
}

func (q Queries) Len() int {
	return len(q)
}

func (q Queries) Less(i, j int) bool {
	return q[i].Order() < q[j].Order()
}

func (q Queries) Swap(i, j int) {
	q[i] , q[j] = q[j], q[i]
}

type ConditionURI string

func (c ConditionURI) URI() string {
	return string(c)
}

func NewConditions() *Conditions {
	return &Conditions{
		limit: 210000000,
		Queries: []Query{},
		cursorQuery: nil,
		cursor: nil,
	}
}

func (w *Conditions) CollectionGroup(collectionId string) *Conditions {
	w.group = collectionId
	return w
}

func (w *Conditions) CollectionGroupWithKey(key *Key) *Conditions {
	return w.CollectionGroup(key.Collection)
}

func (w *Conditions) Where(column string, operation string, value interface{}) *Conditions {
	w.Queries = append(w.Queries, Where{column, operation, value})
	return w
}

func (w *Conditions) Limit(limit int) *Conditions {
	w.limit = limit
	w.Queries = append(w.Queries, Limit(limit))
	return w
}

func (w *Conditions) Offset(offset int) *Conditions {
	w.Queries = append(w.Queries, Offset(offset))
	return w
}

func (w *Conditions) OrderBy(column string, direction firestore.Direction) *Conditions {
	w.Queries = append(w.Queries, Order{column,direction})
	if w.cursor == nil {
		w.cursor = newCursor()
	}
	w.cursor.AddField(column, direction)
	return w
}

func (w *Conditions) StartAfter(cursor *Cursor) *Conditions {
	w.cursorQuery = &StartAfter{}
	w.cursor = cursor
	return w
}

// TODO: StartAt, EndsAt, EndsAfter ...etc

type StartAfter struct {
}

func (w StartAfter) Queryer(query firestore.Query, cursor *Cursor, f *Foon) (firestore.Query, error) {
	doc, err := cursor.snapshot(f)
	if err != nil {
		return query, err
	}
	query = cursor.setOrders(query)
	query = query.StartAfter(doc)
	return query, nil
}

func (w StartAfter) Hash(cursor *Cursor) string {
	if cursor == nil {
		return "startAfter"
	}
	return fmt.Sprintf("startAftr:%s", cursor.planeCursor())
}


type Where struct {
	Column string
	Operation string
	Value interface{}
}

func (w Where) Queryer(query firestore.Query) firestore.Query {
	return query.Where(w.Column, w.Operation, w.Value)
}

func (w Where) Hash() string {
	return fmt.Sprintf("%s-%s-%v", w.Column, w.Operation, w.Value)
}

func (w Where) Order() int {
	length := len(w.Column)
	if length == 0 {
		return 0
	}
	if length == 1 {
		return int(w.Column[0])
	}
	return int(w.Column[0]) * 10000 +  int(w.Column[1])
}

type Offset int

func (w Offset) Queryer(query firestore.Query) firestore.Query {
	return query.Offset(int(w))
}

func (w Offset) Hash() string {
	return fmt.Sprintf("offset%d", w)
}

func (w Offset) Order() int {
	return 10000000
}

type Limit int

func (w Limit) Queryer(query firestore.Query) firestore.Query {
	return query.Limit(int(w))
}

func (w Limit) Hash() string {
	return fmt.Sprintf("limit%d", w)
}

func (w Limit) Order() int {
	return 10000100
}

type Order struct {
	Column string
	Direction firestore.Direction
}

func (w Order) Queryer(query firestore.Query) firestore.Query {
	return query.OrderBy(w.Column, w.Direction)
}

func (w Order) Hash() string {
	return fmt.Sprintf("orderBy:%d-%v", w.Column, w.Direction)
}

func (w Order) Order() int {
	return 10000200
}

func (c Conditions) HasConditions() bool {
	return len(c.Queries) > 0
}

func (c Conditions) HasNoConditions() bool {
	return len(c.Queries) == 0
}

func (c Conditions) Hash() string {
	if len(c.Queries) == 0 {
		return ""
	}
	sort.Sort(c.Queries)

	buf := bytes.Buffer{}
	if c.group != "" {
		buf.WriteString(c.group)
	}

	for _, cond := range c.Queries {
		buf.WriteString(cond.Hash())
	}

	if c.cursorQuery != nil {
		buf.WriteString(c.cursorQuery.Hash(c.cursor))
	}

	hash := md5.New()
	hash.Write(buf.Bytes())
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (c Conditions) String() string {
	if len(c.Queries) == 0 {
		return ""
	}
	sort.Sort(c.Queries)
	buf := bytes.Buffer{}
	for _, cond := range c.Queries {
		buf.WriteString(cond.Hash())
		buf.WriteString("\n")
	}
	return buf.String()
}

func (c Conditions) Query(query firestore.Query, f *Foon) (firestore.Query, error) {
	return c.query(query, f)
}

func (c Conditions) query(query firestore.Query, f *Foon) (firestore.Query, error) {
	for _ , q := range c.Queries {
		query = q.Queryer(query)
	}
	if c.cursor != nil && c.cursorQuery != nil {
		q, err := c.cursorQuery.Queryer(query, c.cursor, f)
		if err != nil {
			return query, err
		}
		query = q
	}
	return query , nil
}

func (c Conditions) URI(key *Key) IURI {
	if c.HasNoConditions() {
		return CollectionCache.CreateURIByKey(key)
	}
	return ConditionURI(fmt.Sprintf("foon/%s/conds/%s", key.CollectionPath(), c.Hash()))
}

func (c Conditions) CursorURI(key *Key) IURI {
	uri := c.URI(key)
	return ConditionURI(fmt.Sprintf("%s/%s", uri.URI(), "cursor"))
}
