package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    input := scanner.Text()

	// the delimiter might be changed
    words := strings.Split(input, ",")
    if len(words) >= 9 {
        fmt.Printf("%s\n", strings.TrimSpace(words[8]))
    } else {
        fmt.Println("Input does not contain enough words.")
    }
}