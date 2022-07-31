package controllers

import (
	"goAuth/database"
	"goAuth/models"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

const secret_key = "secret"

func Register(c *fiber.Ctx) error {
	data := new(models.User)

	if err := c.BodyParser(data); err != nil {
		return err
	}

	pass := []byte(data.Password)
	password, _ := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	user := models.User{
		Name:     data.Name,
		Email:    data.Email,
		Password: string(password),
	}

	database.DB.Create(&user)

	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	data := new(models.User)

	if err := c.BodyParser(data); err != nil {
		return err
	}

	var user models.User

	database.DB.First(&user, "email = ?", data.Email)

	if user.ID == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "user not found",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "incorrect password",
		})
	}

	expiringDate := time.Now().Add(time.Hour + 24).Unix()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.ID)),
		ExpiresAt: jwt.NewTime(float64(expiringDate)),
	})

	token, err := claims.SignedString([]byte(secret_key))
	if err != nil {
		c.Status((fiber.StatusInternalServerError))
		return c.JSON(fiber.Map{
			"message": "error logging in",
		})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)
	return c.JSON(token)
}

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret_key), nil
	})
	if err != nil {
		c.Status((fiber.StatusUnauthorized))
		return c.JSON(fiber.Map{
			"message": "unauthenticated",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	database.DB.First(&user, "ID=?", claims.Issuer)

	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:    "jwt",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}
