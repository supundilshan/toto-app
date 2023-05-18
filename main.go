package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type todoUpdate struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// var todos = []todo{
// 	{ID: "1", Title: "Clean house", Completed: false},
// 	{ID: "2", Title: "Go to gym", Completed: false},
// 	{ID: "3", Title: "Pay bill", Completed: false},
// }

func mongoConnection() (*mongo.Collection, error) {

	//Set up connection
	conString := options.Client().ApplyURI("mongodb://localhost:27017")

	//Connect to mongoDB
	connection, conErr := mongo.Connect(context.Background(), conString)
	if conErr != nil {
		return nil, conErr
	} else {
		fmt.Println(conErr)
	}

	fmt.Println("Database Connected")

	//Connect to collection in mongo database
	collection := connection.Database("go-todo-app").Collection("todos")

	return collection, nil
}

func addTodo(context *gin.Context) {
	var newTodo todo

	// Get mongoDb connection
	mongoDb, mongoErr := mongoConnection()
	if mongoErr != nil {
		return
	} else {
		fmt.Println(mongoErr)
	}

	// Convert json object into Go struct
	if err := context.BindJSON(&newTodo); err != nil {
		fmt.Println(err)
		return
	}

	insertTodo, insErr := mongoDb.InsertOne(context.Request.Context(), newTodo)
	if insErr != nil {
		context.IndentedJSON(http.StatusNotModified, newTodo)
	}

	context.IndentedJSON(http.StatusCreated, insertTodo)
}

func getTodos(context *gin.Context) {
	// Get mongoDb connection
	mongoDb, mongoErr := mongoConnection()
	if mongoErr != nil {
		fmt.Println("mongoErr", mongoErr)
		return
	}

	filter := bson.M{}
	dbData, resErr := mongoDb.Find(context.Request.Context(), filter)
	if resErr == nil {
		fmt.Print("resErr", resErr)
		// context.IndentedJSON(http.StatusNotFound, nil)
	}

	// Iterate
	// var result []todo
	var result []bson.M
	for dbData.Next(context.Request.Context()) {
		var item bson.M
		if err := dbData.Decode(&item); err != nil {
			fmt.Println("err", err)
			context.IndentedJSON(http.StatusNotFound, nil)
		}
		result = append(result, item)
	}

	context.IndentedJSON(http.StatusOK, result)
}

func getTodoById(id string, context *gin.Context) (*todo, error) {
	var result todo
	// Get mongoDb connection
	mongoDb, mongoErr := mongoConnection()
	if mongoErr != nil {
		fmt.Println(mongoErr)
		return nil, nil
	}

	filter := bson.M{"id": id}
	resErr := mongoDb.FindOne(context.Request.Context(), filter).Decode(&result)
	if resErr != nil {
		return nil, resErr
	}
	return &result, nil
}

func getTodo(context *gin.Context) {

	id := context.Param("id")
	todo, err := getTodoById(id, context)

	if err != nil {
		fmt.Print(err)
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, todo)
}

func updateTodo(context *gin.Context) {
	// Get mongoDb connection
	mongoDb, mongoErr := mongoConnection()
	if mongoErr != nil {
		fmt.Println(mongoErr)
	}

	id := context.Param("id")
	filter := bson.M{"id": id}

	var updateValue todoUpdate
	// Convert json object into Go struct
	if err := context.BindJSON(&updateValue); err != nil {
		fmt.Println(err)
	}

	// Find the value is excisting or not
	todo, err := getTodoById(id, context)
	if err != nil {
		fmt.Println("Todo not found", err)
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
	} else {
		println("todo", todo)

		updateTodo := bson.M{"$set": bson.M{"title": updateValue.Title, "completed": updateValue.Completed}}
		updateResult, updateErr := mongoDb.UpdateOne(context.Request.Context(), filter, updateTodo)
		if updateErr != nil {
			fmt.Println(updateErr)
			context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		}

		fmt.Println("updateResult", updateResult)

		context.IndentedJSON(http.StatusOK, updateResult)
	}
}

func deleteTodo(context *gin.Context) {
	// Get mongoDb connection
	mongoDb, mongoErr := mongoConnection()
	if mongoErr != nil {
		fmt.Println(mongoErr)
	}

	id := context.Param("id")
	filter := bson.M{"id": id}

	// Delete a document
	deleteResult, err := mongoDb.DeleteOne(context.Request.Context(), filter)
	if err != nil {
		fmt.Println(err)
	}

	context.IndentedJSON(http.StatusOK, deleteResult.DeletedCount)

	// fmt.Printf("Deleted %v documents\n", deleteResult.DeletedCount)
}

func main() {
	fmt.Print("This is todo app")

	server := gin.Default()

	server.POST("/todos", addTodo)

	server.GET("/todos", getTodos)
	server.GET("/todos/:id", getTodo)

	server.PATCH("/todos/:id", updateTodo)

	server.DELETE("/todos/:id", deleteTodo)

	server.Run("localhost:9090")
}
