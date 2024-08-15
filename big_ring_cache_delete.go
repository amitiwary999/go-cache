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

var bucketNo = 1

type bucket map[string]byte
type deleteInfo struct {
	fileName       string
	tempFileName   string
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

func NewDeleteInfo(hour int, interval time.Duration, homeDir string, fileName string) *deleteInfo {
	tempFileName := fmt.Sprintf("%v/%v", homeDir, "big-cache-ring-data.txt")
	di := &deleteInfo{
		fileName:       fileName,
		tempFileName:   tempFileName,
		deleteHour:     hour,
		buckets:        make(map[int]bucket),
		deleteInterval: interval,
		t:              time.NewTimer(getTickerTime(hour, interval)),
	}
	go di.process()
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

func (d *deleteInfo) clear() {
	offsetMap := make(map[uint64]int64)
	delBucketNo := bucketNo
	bucketNo += 1
	bucket, ok := d.buckets[delBucketNo]
	if ok {
		tmpFile, tmpFileErr := createTempFile(d.tempFileName)
		if tmpFileErr != nil {
			return
		}
		file, fileErr := mainFile(d.fileName)
		if fileErr != nil {
			return
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
						return
					}
					keyInt := xxhash.Sum64([]byte(keyString))
					tmpFile.WriteString(string(b))
					offsetMap[keyInt] = offset
				}
			}
		}
	}
}

func (d *deleteInfo) process() {
	d.clear()
}
