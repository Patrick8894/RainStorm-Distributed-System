package cache

import (
    "container/list"
    "fmt"
    "mp3/src/global"
    "net"
    "os"
    "strings"
    "sync"
    "time"
)

// CacheEntry represents a cache entry
type CacheEntry struct {
    filename     string
    lastModified time.Time
    data         []byte
}

var (
    Cache          = make(map[string]*list.Element)
    CacheList      = list.New()
    MaxCacheSize   = 10 // Maximum number of files to cache
    CacheMutex     sync.Mutex
    CacheDirectory = "cache/"
)

// Cache management functions

func GetCacheEntry(filename string) *CacheEntry {
    /*
    Get the cache entry for the given filename.
    */
    elem, ok := Cache[filename]
    if ok {
        // LRU cache: move the entry to the front of the list
        CacheList.MoveToFront(elem)
        return elem.Value.(*CacheEntry)
    } else {
        return nil
    }
}

func addToCache(filename string, data []byte, lastModified time.Time) {
    /*
    Add a new cache entry for the given filename.
    */

    // if the file is already in the cache, update the entry
    if elem, ok := Cache[filename]; ok {
        CacheList.MoveToFront(elem)
        elem.Value.(*CacheEntry).data = data
        elem.Value.(*CacheEntry).lastModified = lastModified
    } else {
        // Check if the cache is full
        if CacheList.Len() >= MaxCacheSize {
            // Remove least recently used item
            back := CacheList.Back()
            if back != nil {
                // delete cache file from disk
                CacheFilePath := CacheDirectory + back.Value.(*CacheEntry).filename
                os.Remove(CacheFilePath)

                CacheList.Remove(back)
                delete(Cache, back.Value.(*CacheEntry).filename)
            }
        }
        entry := &CacheEntry{
            filename:     filename,
            data:         data,
            lastModified: lastModified,
        }
        elem := CacheList.PushFront(entry)
        Cache[filename] = elem
    }
    // Save cache entry to disk
    CacheFilePath := CacheDirectory + filename
    os.WriteFile(CacheFilePath, data, 0644)
}

func DeleteCacheEntry(filename string) {
    /*
    Delete the cache entry & cache file for the given filename.
    */
    if elem, ok := Cache[filename]; ok {
        CacheList.Remove(elem)
        delete(Cache, filename)
        // Remove cache file from disk
        CacheFilePath := CacheDirectory + filename
        os.Remove(CacheFilePath)
    }
}

func CheckCacheValidity(filename string, CachedLastModified time.Time) bool {
    /*
       Check with the server if the cached version is up-to-date.
    */
    candidates := global.FindFileReplicas(filename)
    if len(candidates) == 0 {
        fmt.Println("No candidates found")
        return false
    }

    // Connect to the server
    conn, err := net.Dial("tcp", candidates[0]+":"+global.HDFSPort)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return false
    }
    defer conn.Close()

    // Send the "stat" command with the filename
    command := fmt.Sprintf("stat %s\n", filename)
    _, err = conn.Write([]byte(command))
    if err != nil {
        fmt.Println("Error sending command:", err)
        return false
    }

    // Read the last modified time from the server
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        fmt.Println("Error reading from connection:", err)
        return false
    }
    response := string(buffer[:n])
    parts := strings.SplitN(response, " ", 2)
    if len(parts) < 2 || parts[0] != "LastModified" {
        fmt.Println("Unexpected response from server:", response)
        return false
    }
    lastModifiedStr := strings.TrimSpace(parts[1])
    lastModified, err := time.Parse(time.RFC3339, lastModifiedStr)
    if err != nil {
        fmt.Println("Error parsing last modified time:", err)
        return false
    }

    return !lastModified.After(CachedLastModified)
}
