package main

import (
	"fmt"
	"time"

	"github.com/goodieshq/sweettooth/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	job := database.Job{
		ID:      1234,
		Created: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	fmt.Println(job)

}
