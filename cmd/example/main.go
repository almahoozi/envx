package main

import (
	"fmt"
	"os"
)

func main() {
	if val, ok := os.LookupEnv("VAL"); ok {
		fmt.Println("VAL:", val)
	} else {
		fmt.Println("VAL not set")
	}

	for i, arg := range os.Args {
		fmt.Printf("arg[%d]: %s\n", i, arg)
	}
}
