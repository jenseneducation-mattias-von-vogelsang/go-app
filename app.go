package main

import (
	"context"
	"encoding/json"
	"time"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const dbName = "fat_stonksdb"
const collectionName = "users"
const port = ":3000"

type User struct {
	_id       string `json:"id,omitempty"`
	UserName  string `json:"username,omitempty"`
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty`
	Age       int    `json:"age,omitempty"`
}

func main() {
	app := fiber.New()

	//TODO: lös login med datat i mongoDB
	app.Post("/login", login)

	app.Get("/", accessible)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("secret"),
	}))

	app.Get("/user/:id?", getUser)

	app.Get("/users", getUsers)

	//TODO: lägg till jwttoken vid signup
	app.Post("/user", postUser)

	app.Put("/user/:id", updateUser)

	app.Delete("/user/:id", deleteUser)

	app.Get("/restricted", restricted)

	app.Listen(port)
}

func getUser(c *fiber.Ctx) error {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	var filter bson.M = bson.M{}

	if c.Params("id") != "" {
		id := c.Params("id")
		objID, _ := primitive.ObjectIDFromHex(id)
		filter = bson.M{"_id": objID}
	}

	var results []bson.M
	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	cur.All(context.Background(), &results)

	if results == nil {
		return c.SendStatus(404)
	}

	json, _ := json.Marshal(results)
	return c.Send(json)
}

func getUsers(c *fiber.Ctx) error {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	var filter bson.M = bson.M{}

	var results []bson.M
	cur, err := collection.Find(context.Background(), filter)
	defer cur.Close(context.Background())

	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	cur.All(context.Background(), &results)

	if results == nil {
		return c.SendStatus(404)
	}

	json, _ := json.Marshal(results)
	return c.Send(json)
}

func postUser(c *fiber.Ctx) error {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	var user User
	json.Unmarshal([]byte(c.Body()), &user)

	res, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	response, _ := json.Marshal(res)
	return c.Send(response)
}

func updateUser(c *fiber.Ctx) error {
	collection, err := getMongoDbCollection(dbName, collectionName)
	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}
	var user User
	json.Unmarshal([]byte(c.Body()), &user)

	update := bson.M{
		"$set": user,
	}

	objID, _ := primitive.ObjectIDFromHex(c.Params("id"))
	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)

	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	response, _ := json.Marshal(res)
	return c.Send(response)
}

func deleteUser(c *fiber.Ctx) error {
	collection, err := getMongoDbCollection(dbName, collectionName)

	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	objID, _ := primitive.ObjectIDFromHex(c.Params("id"))
	res, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if err != nil {
		return c.Status(500).Send([]byte(err.Error()))
	}

	jsonResponse, _ := json.Marshal(res)
	return c.Send(jsonResponse)
}

func login(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("pass")

	// Throws Unauthorized error
	if user != "john" || pass != "doe" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func accessible(c *fiber.Ctx) error {
	return c.SendString("Accessible")
}

func restricted(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}
