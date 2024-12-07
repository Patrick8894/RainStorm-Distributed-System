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
    if len(words) >= 4 {
        fmt.Printf("%s, %s\n", strings.TrimSpace(words[2]), strings.TrimSpace(words[3]))
    } else {
        fmt.Println("Input does not contain enough words.")
    }
}