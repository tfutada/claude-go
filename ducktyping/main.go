// Package main demonstrates Go's duck typing (structural typing)
// through interfaces.
//
// Key concept: "If it walks like a duck and quacks like a duck,
// it's a duck" - types don't declare which interfaces they implement,
// they just need to have the right methods.
package main

import "fmt"

// Speaker defines what it means to speak
type Speaker interface {
	Speak() string
}

// Dog is a concrete type
type Dog struct {
	Name string
}

// Speak makes Dog satisfy Speaker implicitly - no "implements" keyword needed
func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

// Cat is another concrete type
type Cat struct {
	Name string
}

// Cat also satisfies Speaker without explicit declaration
func (c Cat) Speak() string {
	return c.Name + " says: Meow!"
}

// Robot demonstrates that ANY type with Speak() works
type Robot struct {
	Model string
}

func (r Robot) Speak() string {
	return r.Model + " says: Beep boop!"
}

// MakeSpeak accepts any Speaker - duck typing in action
func MakeSpeak(s Speaker) {
	fmt.Println(s.Speak())
}

// Example 2: Multiple interfaces
type Walker interface {
	Walk() string
}

type Talker interface {
	Talk() string
}

// Person satisfies BOTH interfaces
type Person struct {
	Name string
}

func (p Person) Walk() string {
	return p.Name + " is walking"
}

func (p Person) Talk() string {
	return p.Name + " is talking"
}

// WalkAndTalk requires both behaviors - interface composition
type WalkAndTalker interface {
	Walker
	Talker
}

func Demonstrate(wt WalkAndTalker) {
	fmt.Println(wt.Walk())
	fmt.Println(wt.Talk())
}

// Example 3: Empty interface - accepts anything
func PrintAnything(v interface{}) {
	fmt.Printf("Type: %T, Value: %v\n", v, v)
}

// Example 4: Type assertion and type switch
func IdentifyAndSpeak(s Speaker) {
	// Type switch - check concrete type
	switch v := s.(type) {
	case Dog:
		fmt.Printf("It's a dog named %s!\n", v.Name)
	case Cat:
		fmt.Printf("It's a cat named %s!\n", v.Name)
	case Robot:
		fmt.Printf("It's a robot model %s!\n", v.Model)
	default:
		fmt.Println("Unknown speaker type")
	}
	fmt.Println(s.Speak())
}

func main() {
	fmt.Println("=== Go Duck Typing Demo ===\n")

	// 1. Basic duck typing
	fmt.Println("1. Basic Duck Typing:")
	fmt.Println("   (Types satisfy interfaces implicitly)")
	dog := Dog{Name: "Rex"}
	cat := Cat{Name: "Whiskers"}
	robot := Robot{Model: "R2D2"}

	// All can be passed to MakeSpeak because they have Speak()
	MakeSpeak(dog)
	MakeSpeak(cat)
	MakeSpeak(robot)

	// 2. Slice of interface type
	fmt.Println("\n2. Polymorphism via Interface Slice:")
	speakers := []Speaker{dog, cat, robot}
	for _, s := range speakers {
		fmt.Printf("   %s\n", s.Speak())
	}

	// 3. Interface composition
	fmt.Println("\n3. Interface Composition:")
	person := Person{Name: "Alice"}
	Demonstrate(person)

	// 4. Empty interface
	fmt.Println("\n4. Empty Interface (any type):")
	PrintAnything(42)
	PrintAnything("hello")
	PrintAnything(dog)
	PrintAnything([]int{1, 2, 3})

	// 5. Type assertion
	fmt.Println("\n5. Type Assertion & Switch:")
	IdentifyAndSpeak(dog)
	IdentifyAndSpeak(robot)

	// 6. Runtime interface check
	fmt.Println("\n6. Runtime Interface Check:")
	var unknown interface{} = dog
	if speaker, ok := unknown.(Speaker); ok {
		fmt.Printf("   It can speak: %s\n", speaker.Speak())
	}

	// Key insight
	fmt.Println("\n=== Key Insight ===")
	fmt.Println("Go's duck typing means:")
	fmt.Println("- No 'implements' keyword")
	fmt.Println("- Types satisfy interfaces automatically")
	fmt.Println("- Decoupled: interface and type can be in different packages")
	fmt.Println("- Flexible: add new types without modifying interface")
}
