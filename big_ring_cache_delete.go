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
	bucketNo      = 1
	tempFilePath  = ""
	mainFilePath  = ""
	oldDeleteFile *os.File
)

type bucket map[string]byte
type deleteInfo struct {
	tempFile       *os.File
	deleteHour     int
	deleteInterval time.Duration
	deleteMin      int
	deleteSec      int
	buckets        map[int]bucket
	t              *time.Timer
	deleteKeyFile  *os.File
}

func getTickerTime(dInt time.Duration, dHour int, dMin int, dSec int) time.Duration {
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), dHour, dMin, dSec, 0, time.Local)
	if nextTick.Before(now) {
		nextTick = nextTick.Add(dInt)
	}
	return time.Until(nextTick)
}

func createDeleteFile() (*os.File, error) {
	deleteFileName := fmt.Sprintf("%v-%v.txt", DeleteKeyFilePrefix, time.Now().UnixMilli())
	deleteFilePath := fmt.Sprintf("%v/%v/%v", HomeDir, DeleteKeyFileDirectory, deleteFileName)
	return os.OpenFile(deleteFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}

func newDeleteInfo(ti *TickerInfo) (*deleteInfo, error) {
	di := &deleteInfo{
		deleteHour:     ti.Hour,
		deleteInterval: ti.Interval,
		deleteMin:      ti.Min,
		deleteSec:      ti.Sec,
		buckets:        make(map[int]bucket),
		t:              time.NewTimer(getTickerTime(ti.Interval, ti.Hour, ti.Min, ti.Sec)),
	}
	tempFilePath = fmt.Sprintf("%v/%v", HomeDir, "big-cache-ring-data-temp.txt")
	mainFilePath = fmt.Sprintf("%v/%v", HomeDir, FileName)
	deleteFile, err := createDeleteFile()
	di.deleteKeyFile = deleteFile
	return di, err
}

func (d *deleteInfo) loadDeleteKeys() {
	b, ok := d.buckets[bucketNo]
	if !ok {
		b = make(bucket)
		d.buckets[bucketNo] = b
	}
	delDir := fmt.Sprintf("%v/%v", HomeDir, DeleteKeyFileDirectory)
	files, err := os.ReadDir(delDir)
	if err == nil && len(files) > 1 {
		deleteFileName := files[1]
		deleteFilePath := fmt.Sprintf("%v/%v", delDir, deleteFileName)
		deleteOldFile, delErr := os.OpenFile(deleteFilePath, os.O_RDONLY, 0644)
		if delErr != nil {
			return
		}
		scanner := bufio.NewScanner(deleteOldFile)
		scanner.Split(splitFunction)
		for scanner.Scan() {
			byt := scanner.Bytes()
			key := string(byt)
			b[key] = byte(1)
		}
	}
}

func (d *deleteInfo) add(key string) {
	b, ok := d.buckets[bucketNo]
	if !ok {
		b = make(bucket)
		d.buckets[bucketNo] = b
	}
	b[key] = byte(1)
	_, seekErr := d.deleteKeyFile.Seek(0, io.SeekEnd)
	if seekErr == nil {
		deleteFileKey := fmt.Sprintf("%v\n", key)
		d.deleteKeyFile.WriteString(deleteFileKey)
	}
}

func createTempFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}

func mainFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDONLY, 0644)
}

func (d *deleteInfo) updateTicker() {
	d.t.Reset(getTickerTime(d.deleteInterval, d.deleteHour, d.deleteMin, d.deleteSec))
}

func (d *deleteInfo) cleanFile() (map[uint64]int64, []string) {
	offsetMap := make(map[uint64]int64)
	keys := make([]string, 10)
	oldDeleteFile = d.deleteKeyFile
	delBucketNo := bucketNo
	bucketNo += 1
	bucket, ok := d.buckets[delBucketNo]
	deleteKeyNewFile, deleteKeyNewFileErr := createDeleteFile()
	if deleteKeyNewFileErr == nil {
		d.deleteKeyFile = deleteKeyNewFile
	}
	if ok {
		file, fileErr := mainFile(mainFilePath)
		if fileErr != nil {
			fmt.Printf("main file create error %v \n", fileErr)
			return nil, nil
		}
		tmpFile, tmpFileErr := createTempFile(tempFilePath)
		d.tempFile = tmpFile
		if tmpFileErr != nil {
			fmt.Printf("temp file create error %v \n", tmpFileErr)
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

func (d *deleteInfo) clear() {
	d.t.Stop()
	d.tempFile.Close()
}

func (d *deleteInfo) process(intrf CleanFileInterface) {
	for {
		<-d.t.C
		offsetMap, keys := d.cleanFile()
		if offsetMap != nil {
			renameErr := os.Rename(tempFilePath, mainFilePath)
			if renameErr != nil {
				fmt.Printf("failed to rename the temp file after cleanup %v \n", renameErr)
			} else {
				intrf.updateCleanedFile(offsetMap, keys)
			}
		}
		delete(d.buckets, bucketNo-1)
		d.tempFile.Close()
		oldDeleteFile.Close()
		delDir := fmt.Sprintf("%v/%v", HomeDir, DeleteKeyFileDirectory)
		files, err := os.ReadDir(delDir)
		if err == nil {
			for i := 0; i < len(files)-1; i++ {
				os.Remove(files[i].Name())
			}
		}
		d.updateTicker()
	}
}
