/**
 * Created with IntelliJ IDEA.
 * User: Tony
 * Date: 13-10-31
 * Time: 下午3:01
 * To change this template use File | Settings | File Templates.
 */
package ostat

import (
	"github.com/garyburd/redigo/redis"
)

func MakeConn() (redis.Conn, error) {
	return redis.Dial("tcp", "localhost:6379")
}
