package server

import (
	"log"
	"github.com/twinj/uuid"
)

func NewUUID() string {
	id := uuid.NewV4()

	if id == nil {
		log.Fatal("Failed to generate UUID")
	}

	return id.String()
}
