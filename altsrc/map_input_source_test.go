package altsrc

import (
	"testing"
	"time"
)

func TestMapDuration(t *testing.T) {
	inputSource := NewMapInputSource(
		"test",
		map[interface{}]interface{}{
			"duration_of_duration_type": time.Minute,
			"duration_of_string_type":   "1m",
			"duration_of_int_type":      1000,
		})
	d, err := inputSource.Duration("duration_of_duration_type")
	expect(t, time.Minute, d)
	expect(t, nil, err)
	d, err = inputSource.Duration("duration_of_string_type")
	expect(t, time.Minute, d)
	expect(t, nil, err)
	_, err = inputSource.Duration("duration_of_int_type")
	refute(t, nil, err)
}
