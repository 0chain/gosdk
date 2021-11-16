package main

import "net/http"

func main() {

	http.Handle("/", http.FileServer(http.Dir(".")))

	//http.ServeFile(w http.ResponseWriter, r *http.Request, name string)
	http.ListenAndServe(":8090", nil)
}
