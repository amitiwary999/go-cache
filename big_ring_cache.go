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

	bbloom "github.com/amitiwary999/go-cache/internal/bloom"
	xxhash "github.com/cespare/xxhash/v2"
)

var (
	HomeDir                string = ""
	FileName               string = "big-cache-ring-data.txt"
	DeleteKeyFileDirectory string = "delete-key-dir"
	DeleteKeyFilePrefix    string = "delete-key-file"
	DeleteKeyFile          string = ""
)

type bigCacheRing struct {
	file        *os.File
	offsetMap   map[uint64]int64
	cacheRing   *cacheRing[string]
	bloomFilter *bbloom.Bloom
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
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	deleteFileDirErr := os.Mkdir(HomeDir+"/"+DeleteKeyFileDirectory, 0744)
	if deleteFileDirErr != nil && !os.IsExist(deleteFileDirErr) {
		return nil, errors.New("failed to create directory that contain delete key files")
	}
	cacheRing := NewCacheRing[string](bufferSize)
	offsetMap := make(map[uint64]int64)
	filter := bbloom.NewBloomFilter(1000000, 0.01)
	di, deleteFileInfoErr := newDeleteInfo(ti)
	if deleteFileInfoErr != nil {
		fmt.Printf("error in delete info initialization %v \n", deleteFileInfoErr)
		return nil, errors.New("error is delete info initialization")
	}
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

func (c *bigCacheRing) Set(key string, value string) error {
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
	c.bloomFilter.Add(keyInt)
	return nil
}

func (c *bigCacheRing) Get(key string) (string, error) {
	keyByte := []byte(key)
	keyInt := xxhash.Sum64(keyByte)
	itemValue, fetchErr := c.cacheRing.Get(key)
	if fetchErr == nil {
		return itemValue, nil
	} else {
		if !c.bloomFilter.Has(keyInt) {
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
				c.bloomFilter.Add(keyInt)
			}
		}
		offset += int64(len(b))
	}
	tempFilePath = fmt.Sprintf("%v/%v", HomeDir, "big-cache-ring-data-temp.txt")
	_, err := os.Stat(tempFilePath)
	if err == nil {
		os.Remove(tempFilePath)
	}
	c.deleteInfo.loadDeleteKeys()
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
		c.bloomFilter.Clear()
		for _, key := range keys {
			keyByte := []byte(key)
			keyInt := xxhash.Sum64(keyByte)
			c.bloomFilter.Add(keyInt)
		}
	}
}

func (c *bigCacheRing) Clear() {
	c.file.Close()
	c.deleteInfo.clear()
	c.bloomFilter.Clear()
	c.offsetMap = make(map[uint64]int64)
}
