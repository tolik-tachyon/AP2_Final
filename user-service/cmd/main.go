package main

import (
	"fmt"

	"github.com/tolik-tachyon/AP2_Final/user-service/internal/config"
)

func main() {
	fmt.Printf("USER: Test: %q\n", config.Test())
}
