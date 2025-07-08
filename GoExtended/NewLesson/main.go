package main

import (
	"fmt"
	"net/http"
)

func home_page(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "My first page")

}

func contacts_page(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Second Page")
}

func handleRequest() {

	http.HandleFunc("/", home_page)
	http.HandleFunc("/contacts/", contacts_page)
	http.ListenAndServe(":3030", nil)
}

func main() {

	handleRequest()

}
