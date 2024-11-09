package cache

import (
    "container/list"
    "fmt"
    "os"
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
    // CacheDirectory = "cache/"
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

func AddToCache(filename string, localfilename string) {
    /*
    Add a new cache entry for the given filename.
    */
    data, err := os.ReadFile(localfilename)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return
    }

    // if the file is already in the cache, update the entry
    if elem, ok := Cache[filename]; ok {
        CacheList.MoveToFront(elem)
        elem.Value.(*CacheEntry).data = data
        elem.Value.(*CacheEntry).lastModified = time.Now()
    } else {
        // Check if the cache is full
        if CacheList.Len() >= MaxCacheSize {
            // Remove least recently used item
            back := CacheList.Back()
            if back != nil {
                // delete cache file from disk
                // CacheFilePath := CacheDirectory + back.Value.(*CacheEntry).filename
                // os.Remove(CacheFilePath)

                CacheList.Remove(back)
                delete(Cache, back.Value.(*CacheEntry).filename)
            }
        }
        entry := &CacheEntry{
            filename:     filename,
            data:         data,
            lastModified: time.Now(),
        }
        elem := CacheList.PushFront(entry)
        Cache[filename] = elem
    }
    // Save cache entry to disk
    // CacheFilePath := CacheDirectory + filename
    // os.WriteFile(CacheFilePath, data, 0644)
}

func DeleteCacheEntry(filename string) {
    /*
    Delete the cache entry & cache file for the given filename.
    */
    if elem, ok := Cache[filename]; ok {
        CacheList.Remove(elem)
        delete(Cache, filename)
        // Remove cache file from disk
        // CacheFilePath := CacheDirectory + filename
        // os.Remove(CacheFilePath)
    }
}

func CheckCacheValidity(CachedLastModified time.Time) bool {
    /*
       Check with the server if the cached version is up-to-date.
    */

    // Compare current time and CachedLastModified time
    if time.Since(CachedLastModified) < 5*time.Minute {
        return true
    } else {
        fmt.Println("Cached version is stale")
        return false
    }
}
