package cache

import "time"

var bucketNo = 1

const interval time.Duration = 24 * time.Hour

type bucket map[string]byte
type deleteInfo struct {
	deleteHour int
	buckets    map[int]bucket
	t          *time.Timer
}

func getTickerTime(hour int) time.Duration {
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local)
	if nextTick.Before(now) {
		nextTick = nextTick.Add(interval)
	}
	return time.Until(nextTick)
}

func NewDeleteInfo(hour int) *deleteInfo {
	di := &deleteInfo{
		deleteHour: hour,
		buckets:    make(map[int]bucket),
		t:          time.NewTimer(getTickerTime(hour)),
	}
	return di
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
	d.t.Reset(getTickerTime(d.deleteHour))
}

func (d *deleteInfo) process() {

}
