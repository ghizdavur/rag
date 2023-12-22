// repositories.go

package repositories

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `gorm:"unique;not null;size:50" json:"username,omitempty"`
	Passwd    string `gorm:"not null;size:255" json:"passwd,omitempty"`
	FirstName string `gorm:"not null;size:50" json:"first_name,omitempty"`
	LastName  string `gorm:"not null;size:50" json:"last_name,omitempty"`
}

// CreateUser - creates a user in database
func CreateUser(db *gorm.DB, user User) error {
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	fmt.Println("User created successfully!")
	return nil
}

// VerifyUserPassword - this can be use when it's needed to verify the user password
func VerifyUserPassword(db *gorm.DB, username, password string) (bool, error) {
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return false, err
	}

	return checkPasswordHash(password, user.Passwd), nil
}

// BeforeCreate - private method to hash the user's password before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := hashPassword(u.Passwd)
	if err != nil {
		return err
	}
	u.Passwd = hashedPassword
	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/*
	//How to use CreateUser func
	user := repositories.User{
		Username:  "tripframe",
		Passwd:    "12345",
		FirstName: "Ghizdavu",
		LastName:  "Razvan-Marius",
	}

	err := repositories.CreateUser(repositories.DB, user)
	if err != nil {
		log.Fatal(err)
	}
*/
