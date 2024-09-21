package cache

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	xxhash "github.com/cespare/xxhash/v2"
)

var (
	HomeDir  string = ""
	FileName string = "big-cache-ring-data.txt"
)

type bigCacheRing struct {
	file        *os.File
	offsetMap   map[uint64]int64
	cacheRing   *cacheRing[string]
	bloomFilter *bloom.BloomFilter
	deleteInfo  *deleteInfo
}

type TickerInfo struct {
	Hour     int
	Min      int
	Sec      int
	Interval time.Duration
}

type CleanFileInterface interface {
	updateCleanedFile(map[uint64]int64, []string)
}

func NewBigCacheRing(bufferSize int32, ti *TickerInfo) (*bigCacheRing, error) {
	var err error = nil
	HomeDir, err = os.UserHomeDir()
	if err != nil {
		return nil, errors.New("failed to create file")
	}
	filePath := fmt.Sprintf("%v/%v", HomeDir, FileName)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	cacheRing := NewCacheRing[string](bufferSize)
	offsetMap := make(map[uint64]int64)
	filter := bloom.NewWithEstimates(1000000, 0.01)
	di := newDeleteInfo(ti)
	bigch := &bigCacheRing{
		file:        file,
		offsetMap:   offsetMap,
		cacheRing:   cacheRing,
		bloomFilter: filter,
		deleteInfo:  di,
	}
	go di.process(bigch)
	return bigch, nil
}

func (c *bigCacheRing) Save(key string, value string) error {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	offset, err := c.file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("big cache error seeking the file")
		return err
	}
	/** we save key in format {0 or 1}#key. like 1#mykey1 or 0#mykey2. 0# means this key is deleted*/
	fileSaveKey := "1#" + key
	fileData := fileSaveKey + " " + value + "\n"
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
			valueOffset := offset + int64(len(key)) + 2 + 1
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
	offset, ok := c.offsetMap[keyInt]
	if ok {
		_, err := c.file.WriteAt([]byte("0"), offset)
		if err != nil {
			fmt.Printf("failed to update the delete bit %v \n", err)
		}
	}
	delete(c.cacheRing.data, keyInt)
	delete(c.offsetMap, keyInt)
	c.deleteInfo.add(key)
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
			keyDelFlagString := splitStrings[0]
			keyDSplits := strings.Split(keyDelFlagString, "#")
			if len(keyDSplits) > 1 {
				keyString := keyDSplits[1]
				keyInt := xxhash.Sum64([]byte(keyString))
				c.offsetMap[keyInt] = offset
				c.bloomFilter.Add([]byte(keyString))
			}
		}
		offset += int64(len(b))
	}
	tempFilePath = fmt.Sprintf("%v/%v", HomeDir, "big-cache-ring-data-temp.txt")
	_, err := os.Stat(tempFilePath)
	if err == nil {
		os.Remove(tempFilePath)
	}
	done <- 1
}

func (c *bigCacheRing) LoadFileOffset(done chan int) {
	go c.loadFileOffset(done)
}

func (c *bigCacheRing) updateCleanedFile(offsetMap map[uint64]int64, keys []string) {
	if len(keys) > 0 {
		c.file.Close()
		filePath := fmt.Sprintf("%v/%v", HomeDir, FileName)
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("failed to open cleaned file %v \n", err)
		}
		c.file = file
		c.offsetMap = offsetMap
		c.bloomFilter.ClearAll()
		for _, key := range keys {
			c.bloomFilter.Add([]byte(key))
		}
	}
}

func (c *bigCacheRing) Clear() {
	c.file.Close()
	c.deleteInfo.clear()
	c.bloomFilter.ClearAll()
	c.offsetMap = make(map[uint64]int64)
}
