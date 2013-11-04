/**
 * Created with IntelliJ IDEA.
 * User: Tony
 * Date: 13-10-31
 * Time: 下午2:16
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"log"
	"net/http"
	"fmt"
	"encoding/json"
	"time"
	"ostat"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"strconv"
	"strings"
)

type ServerStatus struct {
	Address string
	StartTime time.Time
	StartTimeUnix int64
	Services []string
	Depends []string
}

func main() {
	startTime := time.Now()
	serverStatus := ServerStatus{":8080", startTime, startTime.Unix(), []string {"stat"}, []string {}}

	log.Printf("Starting server at %s...", serverStatus.Address)

	server := &http.Server {
		Addr: serverStatus.Address,
		Handler: goweb.DefaultHttpHandler(),
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	goweb.Map("GET", "/", func (c context.Context) error {
			return JsonResponse(c.HttpResponseWriter(), serverStatus, nil)
		})

	goweb.Map("GET", "/stat_curve/{name}/{period}/increase/{timestamp}", operateWithStatCurveTimestamp(func (c context.Context, curve *ostat.StatCurve, timestamp time.Time) (interface{}, error) {
			return curve.Increase(timestamp)
		}))

	goweb.Map("GET", "/stat_curve/{name}/{period}/increase_by/{timestamp}/{value}", operateWithStatCurveTimestamp(func (c context.Context, curve *ostat.StatCurve, timestamp time.Time) (interface{}, error) {
			strValue := c.PathParams()["value"].(string)
			value, error := strconv.ParseInt(strValue, 10, 64)

			if error != nil {
				return nil, error
			}

			return curve.IncreaseBy(timestamp, int(value))
		}))

	goweb.Map("GET", "/stat_curve/{name}/{period}/get/{timestamp}", operateWithStatCurveTimestamp(func (c context.Context, curve *ostat.StatCurve, timestamp time.Time) (interface{}, error) {
			return curve.GetByTimeStamp(timestamp)
		}))

	goweb.Map("GET", "/stat_curve/{name}/{period}/sequence/{start}/[end]", operateWithStatCurveStartEnd(func (c context.Context, curve *ostat.StatCurve, start, end time.Time) (interface{}, error) {
			return curve.GetSequence(start, end), nil
		}))

	goweb.Map("GET", "/high_order_stat_curve/{name}/{periods}/increase/{timestamp}", operateWithHighOrderStatCurveTimestamp(func (c context.Context, curve *ostat.HighOrderStatCurve, timestamp time.Time) (interface{}, error) {
			return nil, curve.Increase(timestamp)
		}))

	goweb.Map("GET", "/high_order_stat_curve/{name}/{periods}/increase_by/{timestamp}/{value}", operateWithHighOrderStatCurveTimestamp(func (c context.Context, curve *ostat.HighOrderStatCurve, timestamp time.Time) (interface{}, error) {
			strValue := c.PathParams()["value"].(string)
			value, error := strconv.ParseInt(strValue, 10, 64)

			if error != nil {
				return nil, error
			}

			return nil, curve.IncreaseBy(timestamp, int(value))
		}))

	goweb.Map("GET", "/high_order_stat_curve/{name}/{periods}/get/{timestamp}", operateWithHighOrderStatCurveTimestamp(func (c context.Context, curve *ostat.HighOrderStatCurve, timestamp time.Time) (interface{}, error) {
			return curve.Get(timestamp)
		}))

	goweb.Map("GET", "/high_order_stat_curve/{name}/{periods}/sequence/{start}/[end]", operateWithHighOrderStatCurveStartEnd(func (c context.Context, curve *ostat.HighOrderStatCurve, start, end time.Time) (interface{}, error) {
			return curve.GetSequence(start, end)
		}))


	log.Fatal(server.ListenAndServe())
}

type Response struct {
	Result interface{}
	Error error
}

func JsonResponse(w http.ResponseWriter, data interface{}, error error) error {
	response := Response{data, error}
	result, error := json.Marshal(response)

	if error != nil {
		return JsonResponse(w, nil, error)
	} else {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(result))
	}

	return nil
}

func operateWithStatCurve(operation func (context.Context, *ostat.StatCurve) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriod(func (c context.Context, name string, period int) (interface{}, error) {
		return operation(c, ostat.NewStatCurve(name, period))
	})
}

func operateWithStatCurveTimestamp(operation func (context.Context, *ostat.StatCurve, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriodTimestamp(func (c context.Context, name string, period int, timestamp time.Time) (interface{}, error) {
		return operation(c, ostat.NewStatCurve(name, period), timestamp)
	})
}

func operateWithStatCurveStartEnd(operation func (context.Context, *ostat.StatCurve, time.Time, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriodStartEnd(func (c context.Context, name string, period int, start time.Time, end time.Time) (interface{}, error) {
		return operation(c, ostat.NewStatCurve(name, period), start, end)
	})
}

func operateWithHighOrderStatCurve(operation func (context.Context, *ostat.HighOrderStatCurve) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriods(func (c context.Context, name string, periods []int) (interface{}, error) {
		return operation(c, ostat.NewHighOrderStatCurve(name, periods))
	})
}

func operateWithHighOrderStatCurveTimestamp(operation func (context.Context, *ostat.HighOrderStatCurve, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriodsTimestamp(func (c context.Context, name string, periods []int, timestamp time.Time) (interface{}, error) {
		return operation(c, ostat.NewHighOrderStatCurve(name, periods), timestamp)
	})
}

func operateWithHighOrderStatCurveStartEnd(operation func (context.Context, *ostat.HighOrderStatCurve, time.Time, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriodsStartEnd(func (c context.Context, name string, periods []int, start time.Time, end time.Time) (interface{}, error) {
		return operation(c, ostat.NewHighOrderStatCurve(name, periods), start, end)
	})
}

func operateWithNamePeriod(operation func (context.Context, string, int) (interface{}, error)) func (context.Context) error {
	return func(c context.Context) error {
		strName := c.PathParams()["name"].(string)
		strPeriod := c.PathParams()["period"].(string)

		intPeriod, error := strconv.Atoi(strPeriod)

		if error != nil {
			return error
		}

		result, error := operation(c, strName, intPeriod)

		return JsonResponse(c.HttpResponseWriter(), result, error)
	}
}

func operateWithNamePeriodTimestamp(operation func (context.Context, string, int, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriod(func(c context.Context, name string, period int) (interface{}, error) {
		strTimestamp := c.PathParams()["timestamp"].(string)
		timestamp, error := parseTimestamp(strTimestamp)

		if error != nil {
			return nil, error
		}

		return operation(c, name, period, timestamp)
	})
}

func operateWithNamePeriodStartEnd(operation func (context.Context, string, int, time.Time, time.Time) (interface{}, error)) func(context.Context) error {
	return operateWithNamePeriod(func(c context.Context, name string, period int) (interface{}, error) {
		strStart := c.PathParams()["start"].(string)
		start, error := parseTimestamp(strStart)

		if error != nil {
			return nil, error
		}

		end := time.Now()

		if c.PathParams().Has("end") {
			end, error = parseTimestamp(c.PathParams()["end"].(string))
			if error != nil {
				return nil, error
			}
		}

		return operation(c, name, period, start, end)
	})

}

func operateWithNamePeriods(operation func (context.Context, string, []int) (interface{}, error)) func (context.Context) error {
	return func(c context.Context) error {
		strName := c.PathParams()["name"].(string)
		strPeriods := c.PathParams()["periods"].(string)

		strPeriodArray := strings.Split(strPeriods, ",")
		periods := make([]int, len(strPeriodArray), len(strPeriodArray))

		for i, strPeriod := range strPeriodArray {
			intPeriod, error := strconv.Atoi(strPeriod)

			if error != nil {
				return error
			}

			periods[i] = intPeriod
		}

		result, error := operation(c, strName, periods)

		return JsonResponse(c.HttpResponseWriter(), result, error)
	}
}

func operateWithNamePeriodsTimestamp(operation func (context.Context, string, []int, time.Time) (interface{}, error)) func (context.Context) error {
	return operateWithNamePeriods(func(c context.Context, name string, periods []int) (interface{}, error) {
		strTimestamp := c.PathParams()["timestamp"].(string)
		timestamp, error := parseTimestamp(strTimestamp)

		if error != nil {
			return nil, error
		}

		return operation(c, name, periods, timestamp)
	})
}

func operateWithNamePeriodsStartEnd(operation func (context.Context, string, []int, time.Time, time.Time) (interface{}, error)) func(context.Context) error {
	return operateWithNamePeriods(func(c context.Context, name string, periods []int) (interface{}, error) {
		strStart := c.PathParams()["start"].(string)
		start, error := parseTimestamp(strStart)

		if error != nil {
			return nil, error
		}

		end := time.Now()

		if c.PathParams().Has("end") {
			end, error = parseTimestamp(c.PathParams()["end"].(string))
			if error != nil {
				return nil, error
			}
		}

		return operation(c, name, periods, start, end)
	})

}
func parseTimestamp(timestamp string) (time.Time, error) {
	unixTime, error := strconv.ParseInt(timestamp, 10, 64)
	if error != nil {
		return time.Now(), error
	}

	return time.Unix(unixTime, 0), nil
}
