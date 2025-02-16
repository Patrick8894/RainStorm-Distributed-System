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
    words := strings.Split(input, "\x00")
    if len(words) >= 9 {
        fmt.Printf("%s", strings.TrimSpace(words[8]))
    } else {
        fmt.Printf("Input does not contain enough words.")
    }
}