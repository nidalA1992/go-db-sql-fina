package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	parcel.Number = id

	// get
	parc, err := store.Get(parcel.Number)
	require.NoError(t, err)

	assert.Equal(t, parc, parcel)

	err = store.SetStatus(parcel.Number, ParcelStatusRegistered)
	require.NoError(t, err)

	// delete
	err = store.Delete(parc.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	parcel := getTestParcel()
	store := NewParcelStore(db)

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	parc, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, newAddress, parc.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	parcel := getTestParcel()
	store := NewParcelStore(db)

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	parcel.Number = id

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(parcel.Number, ParcelStatusDelivered)
	require.NoError(t, err)

	// check
	parc, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parc.Status, ParcelStatusDelivered)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, parcelMap, len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		p, ok := parcelMap[parcel.Number]
		assert.True(t, ok)

		assert.Equal(t, parcel, p)
	}
}
