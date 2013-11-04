/**
 * Created with IntelliJ IDEA.
 * User: Tony
 * Date: 13-10-31
 * Time: 下午2:55
 * To change this template use File | Settings | File Templates.
 */
package ostat

import (
	"time"
	"strconv"
	"errors"
	"github.com/garyburd/redigo/redis"
)

type StatCurve struct {
	Name string
	Period int
	KeyPrefix string
}

func NewStatCurve(name string, period int) *StatCurve {
	return &StatCurve{name, period, "stat/" + name + "/" + strconv.Itoa(period) + "/#"}
}

func (self *StatCurve) Increase(timestamp time.Time) (int, error) {
	return self.IntOperateByTimeStamp(func (conn redis.Conn, key string) (interface{}, error) {
			return conn.Do("INCR", key)
		}, timestamp)
}

func (self *StatCurve) IncreaseBy(timestamp time.Time, value int) (int, error) {
	return self.IntOperateByTimeStamp(func (conn redis.Conn, key string) (interface{}, error) {
			return conn.Do("INCRBY", key, value)
		}, timestamp)
}

func (self *StatCurve) Remove(timestamp time.Time) error {
	_, error := self.OperateByTimeStamp(func (conn redis.Conn, key string) (interface{}, error) {
			return conn.Do("DEL", key)
		}, timestamp)

	return error
}

func (self *StatCurve) GetByTimeStamp(timestamp time.Time) (int, error) {
	return self.GetByPeriod(self.GetPeriodByTimeStamp(timestamp))
}

func (self *StatCurve) GetByPeriod(period int) (int, error) {
	return self.IntOperateByPeriod(func (conn redis.Conn, key string) (interface{}, error) {
			return conn.Do("GET", key)
		}, period)
}

func (self *StatCurve) GetStartPeriodByTimeStamp(start time.Time) (int, error) {
	return self.GetPeriodByTimeStamp(start), nil
}

func (self *StatCurve) GetStartPeriod() (int, error) {
	conn, error := MakeConn()

	if error != nil {
		return 0, error
	}

	keys, error := conn.Do("KEYS", self.KeyPrefix + "*")

	if error != nil {
		return 0, error
	}

	minPeriod := self.GetPeriodByTimeStamp(time.Now())

	for _, key := range keys.([]string) {
		period, error := strconv.Atoi(key[len(self.KeyPrefix):])
		if error != nil {
			return 0, error
		}
		if period < minPeriod {
			minPeriod = period
		}
	}

	return minPeriod, nil
}

type StatCurvePoint struct {
	Time time.Time
	Value int
}

func (self *StatCurve) GetSequence(start time.Time, end time.Time) []StatCurvePoint {
	startPeriod := self.GetPeriodByTimeStamp(start)
	endPeriod := self.GetPeriodByTimeStamp(end)

	totalPeriods := endPeriod - startPeriod + 1

	output := make([]StatCurvePoint, totalPeriods)

	for period := startPeriod; period <= endPeriod; period ++ {
		value, error := self.GetByPeriod(period)

		if error != nil {
			value = 0
		}

		output[period - startPeriod] = StatCurvePoint{self.GetTime(period), value}
	}

	return output
}

func (self *StatCurve) IntOperateByPeriod(operation func(redis.Conn, string) (interface{}, error), period int) (int, error) {
	result, error := self.OperateByPeriod(operation, period)

	if error != nil {
		return 0, error
	}

	switch result.(type) {
	case int64:
		return int(result.(int64)), error
	case []uint8:
		return strconv.Atoi(string(result.([]uint8)))
	}

	return 0, errors.New("UnknownResultType")
}

func (self *StatCurve) IntOperateByTimeStamp(operation func(redis.Conn, string) (interface{}, error), timestamp time.Time) (int, error) {
	result, error := self.OperateByTimeStamp(operation, timestamp)

	if error != nil {
		return 0, error
	}

	return int(result.(int64)), error
}

func (self *StatCurve) OperateByTimeStamp(operation func(redis.Conn, string) (interface{}, error), timestamp time.Time) (interface{}, error) {
	return self.OperateByKey(operation, self.GetKeyByTimeStamp(timestamp))
}

func (self *StatCurve) OperateByPeriod(operation func(redis.Conn, string) (interface{}, error), period int) (interface{}, error) {
	return self.OperateByKey(operation, self.GetKeyByPeriod(period))
}

func (self *StatCurve) OperateByKey(operation func(redis.Conn, string) (interface{}, error), key string) (interface{}, error) {
	conn, error := MakeConn()

	if error != nil {
		return 0, error
	}

	return operation(conn, key)
}

func (self *StatCurve) GetKeyByTimeStamp(timestamp time.Time) string {
	currentPeriod := self.GetPeriodByTimeStamp(timestamp)
	return self.GetKeyByPeriod(currentPeriod)
}

func (self *StatCurve) GetKeyByPeriod(period int) string {
	return self.KeyPrefix + strconv.Itoa(period)
}

func (self *StatCurve) GetPeriodByTimeStamp(timestamp time.Time) int {
	unixTime := timestamp.Unix()
	return int(unixTime/int64(self.Period))
}

func (self *StatCurve) GetTime(period int) time.Time {
	return time.Unix(int64(period) * int64(self.Period), 0)
}
