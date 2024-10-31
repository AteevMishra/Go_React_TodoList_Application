package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)
var todos []Todo // Declare a global slice to hold Todo items

//Get todos List
func getData(w http.ResponseWriter, r *http.Request) {

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
    // Append the new Todo to the todos slice

	todo.ID = len(todos) + 1
    todos = append(todos, todo)

	fmt.Printf("This is TODO List %+v\n", todos)
    // Send a response back
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "Todo added successfully!"}`))
}

//Update the completetion status of any of the todo items in the list
func updateTodo(w http.ResponseWriter, r *http.Request){
	id:= r.PathValue("id")

	for i, todo := range todos{
		if strconv.Itoa(todo.ID) == id{
			todos[i].Completed = !todos[i].Completed
			respJSON, _:= json.Marshal(todos[i])
			w.WriteHeader(200)
			w.Write(respJSON)
			return
		}
	}

	w.WriteHeader(404)
	w.Write([]byte(`{"error": Todo not found}`))
}

//Delete a todo item
func deleteTodo(w http.ResponseWriter, r *http.Request){
	id:= r.PathValue("id")

	for i, todo := range todos{
		if strconv.Itoa(todo.ID) == id{
			todos = append(todos[:i], todos[i+1:]...)

			respJSON, _:= json.Marshal(todo)
			w.WriteHeader(200)
			w.Write(respJSON)
			return
		}
	}

	w.WriteHeader(404)
	w.Write([]byte(`{"error": Todo not found}`))
}

type Todo struct{
	ID int `json:"id"`
	Completed bool `json:"completed"`
	Body string `json:"body"`
}

func main() {

	router := http.NewServeMux()

	//APIs
	//---
	router.HandleFunc("GET /getData", getData)
	//---
	router.HandleFunc("POST /addTodo", addTodo)
	//---
	router.HandleFunc("PUT /updateTodo/{id}", updateTodo)	
	//---
	router.HandleFunc("DELETE /deleteTodo/{id}", deleteTodo)	
	
	//Start Server
	fmt.Println("Server	starting at PORT 4000")
	log.Fatal(http.ListenAndServe(":4000", router))
}