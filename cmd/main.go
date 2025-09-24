package main

import (
	"fmt"
	"os"

	cmd "github.com/walter2310/nvx/internal"
	"github.com/walter2310/nvx/pkg"
)

func main() {

	nvxPath := os.Getenv("USERPROFILE") + `\nvx\versions\current\bin`
	if err := pkg.AddToUserPath(nvxPath); err != nil {
		fmt.Println("Error:", err)
	}

	cmd.RegisterCommands()
	cmd.Execute()
}
