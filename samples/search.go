package main

import (
	"bytes"
	"connpass-go"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	// Make new connpass API client
	c := connpass.NewClient(nil)

	// Define context
	ctx := context.Background()
	ctxTimeout, _ := context.WithTimeout(ctx, time.Second*30)

	// Define search query params
	qp := connpass.QueryParams{
		KeywordsAnd: []string{"golang"},
		Times: []connpass.Time{
			{Year: 2019, Month: 01},
			{Year: 2019, Month: 02},
		},
		Start: 1,
		Order: connpass.QueryOrderStart,
		Count: 2,
	}

	// Execute search
	results, err := c.SearchEvents(ctxTimeout, qp)
	if err != nil {
		fmt.Printf("SearchEvents failed: %v", err)
		panic(err)
	}

	// Stdout as formatted JSON string
	var buf bytes.Buffer
	resultBytes, err := json.Marshal(results)
	if err != nil {
		fmt.Printf("json.Marshal failed: %v", err)
		panic(err)
	}
	if err := json.Indent(&buf, resultBytes, "", "  "); err != nil {
		fmt.Printf("json.Indent failed: %v", err)
		panic(err)
	}
	fmt.Println(buf.String())
}