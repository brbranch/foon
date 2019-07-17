package foon

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
)

type FirestoreClient interface {
	Get(dr *firestore.DocumentRef) (*firestore.DocumentSnapshot, error)
	GetAll(drs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error)
	Create(dr *firestore.DocumentRef, data interface{}) error
	Set(dr *firestore.DocumentRef, data interface{}, opts ...firestore.SetOption) error
	Delete(dr *firestore.DocumentRef, opts ...firestore.Precondition) error
	Update(dr *firestore.DocumentRef, data []firestore.Update, opts ...firestore.Precondition) error
	Documents(q firestore.Query) *firestore.DocumentIterator
	Batch() (*firestore.WriteBatch, error)
	RunTransaction(fn func(ctx context.Context, fs *firestore.Transaction) error, opts ...firestore.TransactionOption) error
	Client() *firestore.Client
}

type FirestoreClientImpl struct {
	ctx context.Context
	client *firestore.Client
}

func (f *FirestoreClientImpl) Get(dr *firestore.DocumentRef) (*firestore.DocumentSnapshot, error) {
	return dr.Get(f.ctx)
}

func (f *FirestoreClientImpl) GetAll(drs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error) {
	return f.client.GetAll(f.ctx, drs)
}

func (f *FirestoreClientImpl) Create(dr *firestore.DocumentRef, data interface{}) error {
	_, err := dr.Create(f.ctx, data)
	return err
}

func (f *FirestoreClientImpl) Set(dr *firestore.DocumentRef, data interface{}, opts ...firestore.SetOption) error {
	_ , err := dr.Set(f.ctx, data, opts...)
	return err
}

func (f *FirestoreClientImpl) Delete(dr *firestore.DocumentRef, opts ...firestore.Precondition) error {
	_ , err := dr.Delete(f.ctx, opts...)
	return err
}

func (f *FirestoreClientImpl) Update(dr *firestore.DocumentRef, data []firestore.Update, opts ...firestore.Precondition) error {
	_ , err := dr.Update(f.ctx, data , opts...)
	return err
}

func (f *FirestoreClientImpl) Documents(q firestore.Query) *firestore.DocumentIterator {
	return q.Documents(f.ctx)
}

func (f *FirestoreClientImpl) Batch() (*firestore.WriteBatch, error) {
	return f.client.Batch(), nil
}

func (f *FirestoreClientImpl) Client() *firestore.Client {
	return f.client
}

func (f *FirestoreClientImpl) RunTransaction(fn func(ctx context.Context, fs *firestore.Transaction) error, opts ...firestore.TransactionOption) error {
	return f.client.RunTransaction(f.ctx, fn, opts...)
}

type FirestoreTransactionClient struct {
	transaction *firestore.Transaction
	client *firestore.Client
}

func (f *FirestoreTransactionClient) Get(dr *firestore.DocumentRef) (*firestore.DocumentSnapshot, error) {
	return f.transaction.Get(dr)
}

func (f *FirestoreTransactionClient) GetAll(drs []*firestore.DocumentRef) ([]*firestore.DocumentSnapshot, error) {
	return f.transaction.GetAll(drs)
}

func (f *FirestoreTransactionClient) Create(dr *firestore.DocumentRef, data interface{}) error {
	return f.transaction.Create(dr, data)
}

func (f *FirestoreTransactionClient) Set(dr *firestore.DocumentRef, data interface{}, opts ...firestore.SetOption) error {
	return f.transaction.Set(dr, data, opts...)
}

func (f *FirestoreTransactionClient) Delete(dr *firestore.DocumentRef, opts ...firestore.Precondition) error {
	return f.transaction.Delete(dr, opts...)
}

func (f *FirestoreTransactionClient) Update(dr *firestore.DocumentRef, data []firestore.Update, opts ...firestore.Precondition) error {
	return f.transaction.Update(dr, data, opts...)
}

func (f *FirestoreTransactionClient) Documents(q firestore.Query) *firestore.DocumentIterator {
	return f.transaction.Documents(q)
}

func (f *FirestoreTransactionClient) Batch() (*firestore.WriteBatch, error) {
	return nil, errors.New("not supported in transactions")
}

func (f *FirestoreTransactionClient) RunTransaction(fn func(ctx context.Context, fs *firestore.Transaction) error, opts ...firestore.TransactionOption) error {
	return errors.New("not supported")
}

func (f *FirestoreTransactionClient) Client() *firestore.Client {
	return f.client
}