package main

import (
	"fmt"
	"net/http"

	httpServer "github.com/VladislavsPerkanuks/Entain-test-task/internal/http"
)

func main() {
	router := httpServer.NewRouter()

	fmt.Println("Starting server on :3000...")
	if err := http.ListenAndServe(":3000", router); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
