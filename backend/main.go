package main

import (
	"fmt"
	"backend/libraries/server"
)

func main() {
	fmt.Println("Starting Backend Service...")

	srv := server.NewServer()
	err := srv.Start(":8080")
	if  err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

}
