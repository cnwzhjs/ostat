/**
 * Created with IntelliJ IDEA.
 * User: Tony
 * Date: 13-10-31
 * Time: 下午4:05
 * To change this template use File | Settings | File Templates.
 */
package ostat

import (
	"time"
)

type HighOrderStatCurve struct {
	Name string
	Curves map[int]*StatCurve
}

func NewHighOrderStatCurve(name string, periods []int) *HighOrderStatCurve {
	output := HighOrderStatCurve{name, map[int]*StatCurve{}}

	for _, period := range periods {
		output.Curves[period] = NewStatCurve(name, period)
	}

	return &output
}

func (self *HighOrderStatCurve) Increase(timestamp time.Time) error {
	_, error := self.ForEach(func (curve *StatCurve) (interface{}, error) {
		return curve.Increase(timestamp)
	})

	return error
}

func (self *HighOrderStatCurve) IncreaseBy(timestamp time.Time, value int) error {
	_, error := self.ForEach(func (curve *StatCurve) (interface{}, error) {
		return curve.IncreaseBy(timestamp, value)
	})

	return error
}

func (self *HighOrderStatCurve) Get(timestamp time.Time) (map[int]int, error)  {
	results, error := self.ForEach(func (curve *StatCurve) (interface{}, error) {
		return curve.GetByTimeStamp(timestamp)
	})

	output := map[int]int{}

	if error != nil {
		return output, error
	}

	for period, value := range results {
		output[period] = value.(int)
	}

	return output, nil
}

func (self *HighOrderStatCurve) GetSequence(start time.Time, end time.Time) (map[int][]StatCurvePoint, error) {
	results, error := self.ForEach(func (curve *StatCurve) (interface {}, error) {
		return curve.GetSequence(start, end), nil
	})

	output := map[int][]StatCurvePoint{}

	if error != nil {
		return output, error
	}

	for period, data := range results {
		output[period] = data.([]StatCurvePoint)
	}

	return output, nil
}

func (self *HighOrderStatCurve) Remove(timestamp time.Time) error {
	_, error := self.ForEach(func (curve *StatCurve) (interface{}, error) {
		return nil, curve.Remove(timestamp)
	})

	return error
}

func (self *HighOrderStatCurve) ForEach(operation func(curve *StatCurve) (interface{}, error)) (map[int]interface{}, error) {
	results := map[int]interface{}{}

	for period, curve := range self.Curves {
		result, error := operation(curve)
		results[period] = result

		if error != nil {
			return results, error
		}
	}

	return results, nil
}
