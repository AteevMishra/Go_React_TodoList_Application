package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct{
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool `json:"completed"`
	Body string `json:"body"`
}

var collection *mongo.Collection
var todos []Todo // Declare a global slice to hold Todo items

//Get todos List
func getData(w http.ResponseWriter, r *http.Request) {

	var todos []Todo
	cursor, err :=collection.Find(context.Background(),bson.M{})
	if err != nil{
		http.Error(w, err.Error(), 400)
		return
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()){
		var todo Todo
		if err := cursor.Decode(&todo); err != nil{
			http.Error(w, err.Error(), 400)
		return
		}

		todos = append(todos, todo)
	}
	

	allTodos, err:= json.Marshal(todos)
	if(err != nil){
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write(allTodos)
}


//Add todo item to the list
func addTodo(w http.ResponseWriter, r *http.Request) {

	var todo Todo
    err := json.NewDecoder(r.Body).Decode(&todo)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
        w.Write([]byte(`{"message": "Invalid request payload"}`))
        return
    }

	fmt.Printf("This is payload %v\n", todo)

	//Return if user has not passed the body of todo item
	if todo.Body == ""{
		w.WriteHeader(200)
		w.Write([]byte(`{"error": "Body field not passed for the todo item"}`))
		return
	}

	//Add todo item to DB
	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "Failed to add todo item"}`))
        return
	}

	// Send a response back
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)
	addedTodo, _ := json.Marshal(todo)
    w.WriteHeader(http.StatusOK)
    w.Write(addedTodo)
}

//Update the completetion status of any of the todo items in the list
func updateTodo(w http.ResponseWriter, r *http.Request){
	id:= r.PathValue("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid todo ID provided !", 400)
		return
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$set":bson.M{"completed":true}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "Failed to add todo item"}`))
        return
	}

	w.WriteHeader(200)
	w.Write([]byte("Todo task successfully updated !!!"))
}

// //Delete a todo item
func deleteTodo(w http.ResponseWriter, r *http.Request){
	id:= r.PathValue("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil{
		http.Error(w, "Invalid todo ID provided !", 400)
		return
	}

	filter := bson.M{"_id": objectId}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(`{"message": "Failed to delete todo item"}`))
        return
	}

	w.WriteHeader(200)
	w.Write([]byte("Todo task successfully deleted !!!"))
}

func main() {

	err:=godotenv.Load(".env")
	if(err != nil){
		log.Fatal("Error loading .env file: ", err)
	}

	//--Connect to MONGODB database--//
	MONGODB_URI := os.Getenv("MONGODB_URI")
	//Creates a new clientOptions , which holds configuration options for the MongoDB client, and also sets the connection URI of DB cluster
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	//Connects to Project cluster, 'client' is a new MongoDB client instance connected to the server.
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err!=nil{
		log.Fatal(err)
	}

	//Disconnect it, once the 'main' function is completed
	defer client.Disconnect(context.Background())

	//Ping sends a ping command to verify that the client is connected, error means -> server is unreachable, or the network fails etc.
	err =client.Ping(context.Background(), nil)
	if err != nil{
		log.Fatal((err))
	}
	fmt.Println("Connected to MONGODB ATLAS")

	collection = client.Database("golang_db").Collection("todos")


	//APIs
	router := http.NewServeMux()
	//---
	router.HandleFunc("GET /getData", getData)
	//---
	router.HandleFunc("POST /addTodo", addTodo)
	//---
	router.HandleFunc("PUT /updateTodo/{id}", updateTodo)	
	// //---
	router.HandleFunc("DELETE /deleteTodo/{id}", deleteTodo)	
	
	//Start Server
	PORT := os.Getenv("PORT")
	if PORT == ""{
		PORT = "4000"
	}
	fmt.Println("Server	starting at PORT 4000")
	log.Fatal(http.ListenAndServe(":4000", router))
}