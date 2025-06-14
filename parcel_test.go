package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// открываем in-memory базу
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// добавляем
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)

	// читаем
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Status, got.Status)
	require.Equal(t, parcel.Address, got.Address)
	require.Equal(t, parcel.CreatedAt, got.CreatedAt)

	// удаляем
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	store := NewParcelStore(db)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	store := NewParcelStore(db)

	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, got.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	store := NewParcelStore(db)

	// три посылки одного клиента
	client := randRange.Intn(10_000_000)
	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	for i := range parcels {
		parcels[i].Client = client
	}

	// добавляем и сохраняем в map
	parcelMap := make(map[int]Parcel, len(parcels))
	for _, p := range parcels {
		id, err := store.Add(p)
		require.NoError(t, err)
		p.Number = id
		parcelMap[id] = p
	}

	// получаем по client
	stored, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, stored, len(parcels))

	// сверяем
	for _, p := range stored {
		want, ok := parcelMap[p.Number]
		require.True(t, ok, "unexpected parcel number %d", p.Number)
		require.Equal(t, want, p)
	}
}
