package cache

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
)

var (
	bucketNo     = 1
	tempFileName = fmt.Sprintf("%v/%v", HomeDir, "big-cache-ring-data.txt")
)

type bucket map[string]byte
type deleteInfo struct {
	tempFile       *os.File
	deleteHour     int
	deleteInterval time.Duration
	buckets        map[int]bucket
	t              *time.Timer
}

func getTickerTime(hour int, interval time.Duration) time.Duration {
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local)
	if nextTick.Before(now) {
		nextTick = nextTick.Add(interval)
	}
	return time.Until(nextTick)
}

func newDeleteInfo(hour int, interval time.Duration) *deleteInfo {
	di := &deleteInfo{
		deleteHour:     hour,
		buckets:        make(map[int]bucket),
		deleteInterval: interval,
		t:              time.NewTimer(getTickerTime(hour, interval)),
	}
	return di
}

func createTempFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}

func mainFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDONLY, 0644)
}

func (d *deleteInfo) add(key string) {
	b, ok := d.buckets[bucketNo]
	if !ok {
		b = make(bucket)
		d.buckets[bucketNo] = b
	}
	b[key] = byte(1)
}

func (d *deleteInfo) updateTicker() {
	d.t.Reset(getTickerTime(d.deleteHour, d.deleteInterval))
}

func (d *deleteInfo) clear() (map[uint64]int64, []string) {
	offsetMap := make(map[uint64]int64)
	keys := make([]string, 10)
	delBucketNo := bucketNo
	bucketNo += 1
	bucket, ok := d.buckets[delBucketNo]
	if ok {
		tmpFile, tmpFileErr := createTempFile(tempFileName)
		d.tempFile = tmpFile
		if tmpFileErr != nil {
			return nil, nil
		}
		file, fileErr := mainFile(FileName)
		if fileErr != nil {
			return nil, nil
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(splitFunction)
		for scanner.Scan() {
			b := scanner.Bytes()
			splitStrings := strings.Split(string(b), " ")
			if len(splitStrings) > 0 {
				keyString := splitStrings[0]
				_, ok := bucket[keyString]
				if !ok {
					offset, offsetErr := tmpFile.Seek(0, io.SeekEnd)
					if offsetErr != nil {
						return nil, nil
					}
					keyInt := xxhash.Sum64([]byte(keyString))
					tmpFile.WriteString(string(b))
					offsetMap[keyInt] = offset
					keys = append(keys, keyString)
				}
			}
		}
	}
	return offsetMap, keys
}

func (d *deleteInfo) process(intrf CleanFileInterface) {
	for {
		<-d.t.C
		offsetMap, keys := d.clear()
		intrf.updateCleanedFile(offsetMap, keys)
		d.updateTicker()
	}
}
