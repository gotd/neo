package main

import (
	"fmt"
	"time"

	"gortc.io/neo"
)

func main() {
	// Set to current time.
	t := neo.NewTime(time.Now())

	// Travel to future.
	fmt.Println(t.Travel(time.Hour * 2).Format(time.RFC3339))
	// 2019-07-19T16:42:09+03:00

	// Back to past.
	fmt.Println(t.Travel(time.Hour * -2).Format(time.RFC3339))
	// 2019-07-19T14:42:09+03:00
}
