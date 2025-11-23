// Package main demonstrates how duck typing makes testing easier
// compared to traditional OOP approaches.
package main

import "fmt"

// === Production Code ===

// UserRepository - we define the interface we NEED (consumer-side)
type UserRepository interface {
	GetUser(id int) (User, error)
	SaveUser(user User) error
}

type User struct {
	ID   int
	Name string
}

// UserService depends on interface, not concrete type
type UserService struct {
	repo UserRepository
}

func (s *UserService) GetUserName(id int) (string, error) {
	user, err := s.repo.GetUser(id)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (s *UserService) RenameUser(id int, newName string) error {
	user, err := s.repo.GetUser(id)
	if err != nil {
		return err
	}
	user.Name = newName
	return s.repo.SaveUser(user)
}

// === Production Implementation ===

// PostgresUserRepository - real database implementation
type PostgresUserRepository struct {
	connectionString string
}

func (r *PostgresUserRepository) GetUser(id int) (User, error) {
	// Real: query database
	// SELECT * FROM users WHERE id = ?
	fmt.Println("  [DB] Querying PostgreSQL...")
	return User{ID: id, Name: "Real User"}, nil
}

func (r *PostgresUserRepository) SaveUser(user User) error {
	// Real: update database
	// UPDATE users SET name = ? WHERE id = ?
	fmt.Println("  [DB] Saving to PostgreSQL...")
	return nil
}

// === Test Implementation ===

// MockUserRepository - simple mock, no framework needed!
type MockUserRepository struct {
	users       map[int]User
	getCallCount int
	saveCallCount int
}

func NewMockRepo() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[int]User),
	}
}

func (m *MockUserRepository) GetUser(id int) (User, error) {
	m.getCallCount++
	user, ok := m.users[id]
	if !ok {
		return User{}, fmt.Errorf("user %d not found", id)
	}
	return user, nil
}

func (m *MockUserRepository) SaveUser(user User) error {
	m.saveCallCount++
	m.users[user.ID] = user
	return nil
}

// === Why This Is Easier Than Traditional OOP ===

func main() {
	fmt.Println("=== Duck Typing Makes Testing Easier ===\n")

	// 1. Production usage
	fmt.Println("1. Production (real database):")
	prodRepo := &PostgresUserRepository{connectionString: "postgres://..."}
	prodService := &UserService{repo: prodRepo}
	name, _ := prodService.GetUserName(1)
	fmt.Printf("   Got user: %s\n", name)

	// 2. Test usage - no mocking framework needed!
	fmt.Println("\n2. Testing (mock repository):")
	mockRepo := NewMockRepo()
	mockRepo.users[1] = User{ID: 1, Name: "Test User"}

	testService := &UserService{repo: mockRepo}

	// Test GetUserName
	name, err := testService.GetUserName(1)
	if err != nil || name != "Test User" {
		fmt.Println("   FAIL: GetUserName")
	} else {
		fmt.Println("   PASS: GetUserName returned 'Test User'")
	}

	// Test RenameUser
	err = testService.RenameUser(1, "New Name")
	if err != nil || mockRepo.users[1].Name != "New Name" {
		fmt.Println("   FAIL: RenameUser")
	} else {
		fmt.Println("   PASS: RenameUser changed name to 'New Name'")
	}

	// Verify call counts
	fmt.Printf("   GetUser called: %d times\n", mockRepo.getCallCount)
	fmt.Printf("   SaveUser called: %d times\n", mockRepo.saveCallCount)

	// 3. The comparison
	fmt.Println("\n=== Traditional OOP (Java) ===")
	fmt.Println("```java")
	fmt.Println("// 1. Interface must exist FIRST")
	fmt.Println("interface UserRepository { ... }")
	fmt.Println("")
	fmt.Println("// 2. Implementation DECLARES it implements")
	fmt.Println("class PostgresRepo implements UserRepository { ... }")
	fmt.Println("")
	fmt.Println("// 3. For testing, need:")
	fmt.Println("//    - Mockito framework")
	fmt.Println("//    - @Mock annotations")
	fmt.Println("//    - when().thenReturn() setup")
	fmt.Println("@Mock UserRepository mockRepo;")
	fmt.Println("when(mockRepo.getUser(1)).thenReturn(testUser);")
	fmt.Println("```")

	fmt.Println("\n=== Go Duck Typing ===")
	fmt.Println("```go")
	fmt.Println("// 1. Define interface WHERE YOU NEED IT")
	fmt.Println("type UserRepository interface { ... }")
	fmt.Println("")
	fmt.Println("// 2. Any struct with matching methods works")
	fmt.Println("type MockRepo struct { users map[int]User }")
	fmt.Println("func (m *MockRepo) GetUser(id int) (User, error) { ... }")
	fmt.Println("")
	fmt.Println("// 3. No framework, no annotations, no magic")
	fmt.Println("mock := &MockRepo{users: testData}")
	fmt.Println("service := &UserService{repo: mock}")
	fmt.Println("```")

	fmt.Println("\n=== Key Benefits ===")
	fmt.Println("1. No mocking framework needed (Mockito, etc.)")
	fmt.Println("2. Mocks are plain Go structs - easy to debug")
	fmt.Println("3. Full control over mock behavior")
	fmt.Println("4. Can add verification (call counts, args)")
	fmt.Println("5. IDE autocomplete works perfectly")
	fmt.Println("6. Compile-time checking of mock methods")
}
