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
	Address  // embedded - methods and fields are promoted
}

func main() {
	employee := Employee{
		Name:     "Alice",
		Age:      30,
		IsRemote: true,
		Address:  Address{City: "Tokyo", Country: "Japan"},
	}

	// Direct access - methods and fields are promoted
	fmt.Println("Name:", employee.Name)
	fmt.Println("Full Address:", employee.FullAddress()) // direct call
	fmt.Println("City:", employee.City)                  // direct access

	// Still works via Address too
	fmt.Println("Via Address:", employee.Address.FullAddress())
}
