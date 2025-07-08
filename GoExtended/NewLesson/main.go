package main

import (
	"fmt"
	"net/http"
)

func home_page(w http.ResponseWriter, r *http.Request) {

}

func main() {

	http.HandleFunc("/", home_page)
	http.ListenAndServe(":8080", nil)

	fmt.Println()

}
