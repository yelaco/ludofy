package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const envFile = "./configs/stofinet/app.env"

func main() {
	// Ask for Stockfish path
	fmt.Print("Enter the path to Stockfish binary: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	stockfishPath := strings.TrimSpace(scanner.Text())

	if stockfishPath == "" {
		fmt.Println("Error: Stockfish path cannot be empty.")
		return
	}

	// Check if Stockfish binary is runnable
	if err := checkStockfish(stockfishPath); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write to app.env
	file, err := os.Create(envFile)
	if err != nil {
		fmt.Println("Error creating .env file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("STOCKFISH_PATH=" + stockfishPath + "\n")
	if err != nil {
		fmt.Println("Error writing to .env file:", err)
		return
	}

	fmt.Println("stockfish path saved successfully to app.env")
}

// checkStockfish verifies if the given Stockfish binary is executable
func checkStockfish(path string) error {
	cmd := exec.Command(path, "uci")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stockfish binary is not runnable: %v", err)
	}
	fmt.Println("stockfish binary is verified and runnable.")
	return nil
}
