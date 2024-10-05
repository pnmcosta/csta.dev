package models

import (
	"time"
)

type Post struct {
	Title   string
	Summary string
	Tags    []string
	Date    time.Time
	Content []byte
}
