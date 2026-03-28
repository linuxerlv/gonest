package main

import (
	"fmt"
	"os"

	"github.com/linuxerlv/gonest/internal/mixingen"
)

func main() {
	code, err := mixingen.GenerateMiddlewareMixin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating mixin: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("mixin_generated.go", code, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mixin code generated successfully: mixin_generated.go")
}
