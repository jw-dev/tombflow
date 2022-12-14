package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jw-dev/tombflow/pkg/script"
)

func fatal(err string) {
	fmt.Printf("%v\nUsage: %v SCRIPT\n", err, os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) == 1 {
		fatal("Not enough arguments")
	}

	path := os.Args[1]
	f, err := os.Open(path)
	if err != nil {
		fatal(fmt.Sprintf("Error opening file: %v", err))
	}
	defer f.Close()

	s, err := script.Read(f)
	if err != nil {
		log.Fatalf("Critical error reading script\n%v\n", err)
	}
	for i, level := range s.Levels {
		fmt.Printf("Level %v: %v (%v)\n", i, level.Name, level.Path)
		fmt.Printf("  Puzzles: %v\n", level.Puzzles)
		fmt.Printf("  Keys: %v\n", level.Keys)
		fmt.Printf("  Pickups: %v\n", level.Pickups)
		fmt.Println("  Flow: ")
		for j, cmd := range level.Flow {
			fmt.Printf("    %v: %v\n", j, s.FormatCommand(cmd))
		}
	}
}
