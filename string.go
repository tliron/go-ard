package ard

import (
	"strconv"
	"time"

	"github.com/tliron/kutil/util"
)

func ValueToString(value Value) string {
	switch value_ := value.(type) {
	case bool:
		return strconv.FormatBool(value_)
	case int64:
		return strconv.FormatInt(value_, 10)
	case int32:
		return strconv.FormatInt(int64(value_), 10)
	case int8:
		return strconv.FormatInt(int64(value_), 10)
	case int:
		return strconv.FormatInt(int64(value_), 10)
	case uint64:
		return strconv.FormatUint(value_, 10)
	case uint32:
		return strconv.FormatUint(uint64(value_), 10)
	case uint8:
		return strconv.FormatUint(uint64(value_), 10)
	case uint:
		return strconv.FormatUint(uint64(value_), 10)
	case float64:
		return strconv.FormatFloat(value_, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(value_), 'g', -1, 32)
	case time.Time:
		return value_.String()
	default:
		return util.ToString(value)
	}
}
