package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

type Todo struct {
    Id       int      `json:"id"`
    Body     string   `json:"body" validate:"required"`  // optional validation
    Completed bool     `json:"completed"`
}

func main() {
    // Initialize a new Fiber app
    app := fiber.New()
    app.Use(cors.New())

    envErr := godotenv.Load(".env")
    if envErr != nil{
        log.Fatal("Error while loading .env")
    }

    PORT := os.Getenv("PORT")
    
    var todos []Todo

    // Define a route for the GET method on the root path '/'
    app.Get("/api/todos", func(c fiber.Ctx) error {
        // Send a string response to the client
        return c.Status(200).JSON(todos)
    })

    app.Post("/api/todos", func(c fiber.Ctx) error {
        todo := &Todo{} // Initialize a new Todo struct
    
        // Parse the request body into the todo struct
        if err := json.Unmarshal(c.Body(), &todo); err != nil {
			return err
		}
    
        // Validate that the Body field is not empty
        if todo.Body == "" {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Todo Body is required"})
        }
    
        // Set the ID for the new todo and append it to the todos slice
        todo.Id = len(todos) + 1
        todos = append(todos, *todo)
    
        // Return the created todo with a 201 status code
        return c.Status(fiber.StatusCreated).JSON(todo)
    })
    
    // Update todo
    app.Patch("/api/todos/:id", func(c fiber.Ctx) error{
        id:= c.Params("id")

        for i, todo := range todos{
            if fmt.Sprint(todo.Id) == id{
                todos[i].Completed = true
                return c.Status(200).JSON(todos[i])
            }
        }

        return c.Status(400).JSON(fiber.Map{"error": "Todo not found!"})
    })


    // delete todos
    app.Delete("/api/todos/:id", func(c fiber.Ctx) error{
        id := c.Params("id")

        for i, todo := range todos{
            if fmt.Sprint(todo.Id) == id{
                todos=append(todos[:i], todos[i+1:]...)
                return c.Status(200).JSON(fiber.Map{"success": "Delete a todo."})
            }
        }

        return c.Status(400).JSON(fiber.Map{"error": "Todo not found!"})
    })

    // Start the server on port 4000
    log.Fatal(app.Listen(PORT))
}