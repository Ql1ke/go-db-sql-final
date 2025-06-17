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
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, id, 0)

	// get
	parcel.Number = id
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, got)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// get and verify
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, got.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// set status
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// get and verify
	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, got.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	client := randRange.Intn(10_000_000)

	// prepare parcels
	parcels := []Parcel{getTestParcel(), getTestParcel(), getTestParcel()}
	parcelMap := make(map[int]Parcel, len(parcels))
	for i := range parcels {
		parcels[i].Client = client
	}

	// add and record
	for _, p := range parcels {
		id, err := store.Add(p)
		require.NoError(t, err)
		p.Number = id
		parcelMap[id] = p
	}

	// get by client
	stored, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, stored, len(parcels))

	// verify each
	for _, p := range stored {
		want, ok := parcelMap[p.Number]
		require.True(t, ok, "unexpected parcel number %d", p.Number)
		require.Equal(t, want, p)
	}
}
