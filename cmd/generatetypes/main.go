// generatetypes is a command-line tool that generates Lua type definitions
// from Go struct types defined in the internal/config package. The type
// definitions are printed to standard output.
package main

import (
	"fmt"
	"log"

	"github.com/RyRose/uplog/internal/config"
)

func main() {
	output, err := config.GenerateLuaTypesFile()
	if err != nil {
		log.Fatalf("Failed to generate types file: %v", err)
	}
	fmt.Println(output)
}
