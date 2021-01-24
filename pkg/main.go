package main

import "fmt"

func main() {
	store := Store{}
	err := store.Open("./myfile.db", "+fwip.myplace")
	if err != nil {
		panic(err)
	}
	defer store.Close()
	workspace, err := store.Workspace()
	if err != nil {
		fmt.Printf("Err! %s\n", err)
	}
	fmt.Println(workspace)
	Serve("hello.db")
}
