package cache

import "time"

var bucketNo = 1

type bucket map[string]byte
type deleteInfo struct {
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

func NewDeleteInfo(hour int, interval time.Duration) *deleteInfo {
	di := &deleteInfo{
		deleteHour:     hour,
		buckets:        make(map[int]bucket),
		deleteInterval: interval,
		t:              time.NewTimer(getTickerTime(hour, interval)),
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
	d.t.Reset(getTickerTime(d.deleteHour, d.deleteInterval))
}

func (d *deleteInfo) clear(bucketNo int) {

}

func (d *deleteInfo) process() {

}
