package cache

import (
	"fmt"
	"io"
	"os"
)

type bigCacheRing struct {
	file      *os.File
	offsetMap map[string]int64
	cacheRing *cacheRing[string]
}

func NewBigCacheRing() (*bigCacheRing, error) {
	file, err := os.OpenFile("big-cache-ring-data.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	cacheRing := NewCacheRing[string](1000)
	offsetMap := make(map[string]int64)
	return &bigCacheRing{
		file:      file,
		offsetMap: offsetMap,
		cacheRing: cacheRing,
	}, nil
}

func (c *bigCacheRing) Save(key string, value string) error {
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("big cache error seeking the file")
		return err
	}
	fileData := key + " " + value
	_, saveErr := c.file.WriteString(fileData)
	if saveErr != nil {
		return saveErr
	}
	c.offsetMap[key] = offset
	return nil
}
