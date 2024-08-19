package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Body      string             `json:"body" validate:"required"` // optional validation
	Completed bool               `json:"completed"`
}

var collection *mongo.Collection

func main() {

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatal("No .env file found", err)
		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("CONNECTED TO MONGO DB")

	collection = client.Database("todo_db").Collection("todos")

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))
	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "4000"
	}

	// if os.Getenv("ENV") == "production" {
	// 	app.Static("/", "./client/dist")
	// }

	log.Fatal(app.Listen("0.0.0.0:" + PORT))
}

func getTodos(c fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo

		if err := cursor.Decode(&todo); err != nil {
			return err
		}

		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func createTodo(c fiber.Ctx) error {
	todo := new(Todo)

	if err := json.Unmarshal(c.Body(), todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body can not be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.Id = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodo(c fiber.Ctx) error {
	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodo(c fiber.Ctx) error {
	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}
