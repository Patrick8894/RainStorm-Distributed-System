package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    // Count the number of words in the input
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    input := scanner.Text()

	// the delimiter might be changed
    words := strings.Split(input, "\0")
    if len(words) >= 9 {
        fmt.Printf("%s\n", strings.TrimSpace(words[8]))
    } else {
        fmt.Println("Input does not contain enough words.")
    }
}