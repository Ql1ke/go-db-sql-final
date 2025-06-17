package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}
	parcel.Number = id

	fmt.Printf(
		"Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt,
	)

	return parcel, nil
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	for _, p := range parcels {
		fmt.Printf(
			"Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			p.Number, p.Address, p.Client, p.CreatedAt, p.Status,
		)
	}
	fmt.Println()
	return nil
}

func (s ParcelService) NextStatus(number int) error {
	p, err := s.store.Get(number)
	if err != nil {
		return err
	}

	var next string
	switch p.Status {
	case ParcelStatusRegistered:
		next = ParcelStatusSent
	case ParcelStatusSent:
		next = ParcelStatusDelivered
	case ParcelStatusDelivered:
		return nil
	}

	fmt.Printf("У посылки № %d новый статус: %s\n", number, next)
	return s.store.SetStatus(number, next)
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"

	p1, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	if err = service.ChangeAddress(p1.Number, newAddress); err != nil {
		fmt.Println(err)
		return
	}

	if err = service.NextStatus(p1.Number); err != nil {
		fmt.Println(err)
		return
	}

	if err = service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}

	_ = service.Delete(p1.Number)

	if err = service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}

	p2, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = service.Delete(p2.Number); err != nil {
		fmt.Println(err)
		return
	}

	if err = service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}
}
