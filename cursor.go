package foon

import (
	"crypto/aes"
		"encoding/base64"
	"cloud.google.com/go/firestore"
	"fmt"
	"strings"
	"errors"
	"strconv"
	"io"
	"crypto/rand"
	cipher "crypto/cipher"
)

/** DocumentSnapshot使いづらいので粒度を上げるためのカーソル作成 */
type Cursor struct {
	ID string
	Path string
	Orders []CursorOrder
}

type CursorOrder struct {
	FieldName string
	Direction firestore.Direction
}

var cursorKey = []byte("2f8472d791c48224ec9753505d60e205")


func newCursor() *Cursor {
	return &Cursor{"", "", []CursorOrder{}}
}

func (c Cursor) snapshot(f *Foon) (*firestore.DocumentSnapshot, error) {
	if c.Path == "" {
		return nil, errors.New("cursor is not defined")
	}
	return f.client.Client().Doc(c.Path).Get(f.Context)
}

func (c Cursor) setOrders(query firestore.Query) firestore.Query {
	for _ , order := range c.Orders {
		query = query.OrderBy(order.FieldName, order.Direction)
	}
	return query
}

func (c Cursor) String() string {
	return c.StringWithSeed(cursorKey)
}

func (c Cursor) planeCursor() string {
	slises := []string{c.ID,c.Path}
	for _, order := range c.Orders {
		slises = append(slises, fmt.Sprintf("%s.%d", order.FieldName, order.Direction))
	}
	return strings.Join(slises, ":")
}

func (c *Cursor) AddField(column string, direction firestore.Direction) {
	c.Orders = append(c.Orders, CursorOrder{column, direction})
}

func (c Cursor) NewCursorWithOrders() *Cursor {
	return &Cursor{
		ID: "",
		Path: "",
		Orders: c.Orders,
	}
}

func (c Cursor) StringWithSeed(seed []byte) string {
	block, err := aes.NewCipher(seed)
	if err != nil {
		return ""
	}
	planeText := []byte(c.planeCursor())

	cipherText := make([]byte, aes.BlockSize+ len(planeText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return ""
	}
	encryptStream := cipher.NewCTR(block, iv)
	encryptStream.XORKeyStream(cipherText[aes.BlockSize:], planeText)

	return base64.URLEncoding.EncodeToString(cipherText)
}

func NewCursor(cursor string) (*Cursor, error) {
	return NewCursorWithSeed(cursor, cursorKey)
}

func NewCursorWithSeed(cursor string, seed []byte) (*Cursor, error) {
	if cursor == "" {
		return nil, nil
	}
	block, err := aes.NewCipher(seed)
	if err != nil {
		return nil, err
	}

	cipherText , err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	decryptedText := make([]byte, len(cipherText[aes.BlockSize:]))
	decryptStream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	decryptStream.XORKeyStream(decryptedText, cipherText[aes.BlockSize:])

	return decodeCursor(string(decryptedText))
}

func decodeCursor(decodeString string) (*Cursor , error){
	slises := strings.Split(decodeString, ":")
	id := slises[0]
	path := slises[1]
	orders := []CursorOrder{}
	if len(slises) > 2 {
		for i := 2; i < len(slises); i++ {
			strs := strings.Split(slises[i], ".")
			if len(strs) != 2 {
				return nil, errors.New("cursor is invalid")
			}
			order, err := strconv.ParseInt(strs[1], 10,64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse order (reason:%v)", err)
			}
			orders = append(orders, CursorOrder{strs[0], firestore.Direction(order)})
		}
	}
	return &Cursor{id, path, orders}, nil
}


