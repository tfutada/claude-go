package main

import "fmt"

type Address struct {
	City    string
	Country string
}

func (a Address) FullAddress() string {
	return a.City + ", " + a.Country
}

type Employee struct {
	Name     string
	Age      int
	IsRemote bool
	Address  Address // named field
}

func main() {
	employee := Employee{
		Name:     "Alice",
		Age:      30,
		IsRemote: true,
		Address:  Address{City: "Tokyo", Country: "Japan"},
	}

	// Must access via field name
	fmt.Println("Name:", employee.Name)
	fmt.Println("Full Address:", employee.Address.FullAddress())
	fmt.Println("City:", employee.Address.City)
}
