package cache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	xxhash "github.com/cespare/xxhash/v2"
)

type bigCacheRing struct {
	file        *os.File
	homeDir     string
	offsetMap   map[uint64]int64
	cacheRing   *cacheRing[string]
	bloomFilter *bloom.BloomFilter
}

func NewBigCacheRing(bufferSize int32) (*bigCacheRing, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("failed to create file")
	}
	fileName := fmt.Sprintf("%v/%v", homeDir, "big-cache-ring-data.txt")
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	cacheRing := NewCacheRing[string](bufferSize)
	offsetMap := make(map[uint64]int64)
	filter := bloom.NewWithEstimates(1000000, 0.01)
	return &bigCacheRing{
		homeDir:     homeDir,
		file:        file,
		offsetMap:   offsetMap,
		cacheRing:   cacheRing,
		bloomFilter: filter,
	}, nil
}

func (c *bigCacheRing) Save(key string, value string) error {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("big cache error seeking the file")
		return err
	}
	fileData := key + " " + value + "\n"
	_, saveErr := c.file.WriteString(fileData)
	if saveErr != nil {
		return saveErr
	}
	c.offsetMap[keyInt] = offset
	c.bloomFilter.Add([]byte(key))
	return nil
}

func (c *bigCacheRing) Get(key string) (string, error) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	itemValue, fetchErr := c.cacheRing.Get(key)
	if fetchErr == nil {
		return itemValue, nil
	} else {
		if !c.bloomFilter.Test([]byte(key)) {
			return "", errors.New("key not found")
		}
		offset, ok := c.offsetMap[keyInt]
		if ok {
			valueOffset := offset + int64(len(key)) + 1
			_, fileErr := c.file.Seek(valueOffset, 0)
			if fileErr != nil {
				return "", fileErr
			}
			buffer := make([]byte, 1)
			var content []byte

			for {
				n, err := c.file.Read(buffer)
				if err != nil {
					return "", err
				}

				if n == 0 {
					break
				}

				if buffer[0] == '\n' {
					break
				}

				content = append(content, buffer[0])
			}
			value := string(content)
			c.cacheRing.Set(key, value)
			return string(content), nil
		} else {
			return "", errors.New("key not found")
		}
	}
}

func (c *bigCacheRing) Delete(key string) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	delete(c.cacheRing.data, keyInt)
	delete(c.offsetMap, keyInt)
}

func (c *bigCacheRing) Size(cacheType int) int {
	if cacheType == 1 {
		return len(c.offsetMap)
	} else {
		return len(c.cacheRing.data)
	}
}

func splitFunction(data []byte, eof bool) (next int, token []byte, err error) {
	if eof && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}

	if eof {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (c *bigCacheRing) loadFileOffset(done chan int) {
	scanner := bufio.NewScanner(c.file)
	scanner.Split(splitFunction)
	offset := int64(0)
	for scanner.Scan() {
		b := scanner.Bytes()
		splitStrings := strings.Split(string(b), " ")
		if len(splitStrings) > 0 {
			keyString := splitStrings[0]
			keyInt := xxhash.Sum64([]byte(keyString))
			c.offsetMap[keyInt] = offset
			c.bloomFilter.Add([]byte(keyString))
		}
		offset += int64(len(b))
	}
	done <- 1
}

func (c *bigCacheRing) LoadFileOffset(done chan int) {
	go c.loadFileOffset(done)
}
