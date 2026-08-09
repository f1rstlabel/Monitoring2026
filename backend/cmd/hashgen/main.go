package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	hash := "$2a$12$1uY.YmN67eFj/7dM1h2kEe3Z1gV4L9.Y5E5m/O7j1L0.d1Y.E1/E."
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123"))
	fmt.Println("Matches?", err == nil)
	fmt.Println("Error:", err)
}
