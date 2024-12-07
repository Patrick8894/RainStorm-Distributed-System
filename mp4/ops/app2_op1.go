package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: go run program.go X")
        return
    }

    X := os.Args[1]

    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    input := scanner.Text()

	// the delimiter might be changed
    words := strings.Split(input, ",")
    if len(words) >= 7 && strings.TrimSpace(words[6]) == X {
        fmt.Println(1)
    } else {
        fmt.Println(0)
    }
}