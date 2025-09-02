package main

import (
	"log"
	"os"

	"github.com/massanaRoger/m/v2/internal/kube"
)

func main() {
	provider, err := kube.NewProvider()
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	client, err := provider.ClientSet()

	kube.ListPods(client)
}
