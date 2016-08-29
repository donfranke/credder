package main

import (
	"time"
)

type Page struct {
	Title string
	Body  []byte
}

type Cred struct {
	ID         string
	SecretInfo string
	KeyID      string
}

type LogEvent struct {
	Timestamp        time.Time
	EventDescription string
}

type Key struct {
	ID  string
	Key string
}
