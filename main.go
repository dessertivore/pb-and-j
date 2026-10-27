package main

import (
	"context"
	"fmt"
	"log"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	champs, err := getChamps(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("found %d champions\n", len(champs))
	for _, c := range champs {
		fmt.Printf("%-18s %s\n", c.Name, c.Slug)
	}
}
