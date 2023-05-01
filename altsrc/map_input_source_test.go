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

func TestMapInputSource_Int64Slice(t *testing.T) {
	inputSource := NewMapInputSource(
		"test",
		map[interface{}]interface{}{
			"test_num": []interface{}{int64(1), int64(2), int64(3)},
		})
	d, err := inputSource.Int64Slice("test_num")
	expect(t, []int64{1, 2, 3}, d)
	expect(t, nil, err)
}

func TestMapInputSource_mapToJsonable(t *testing.T) {
	testMap := map[interface{}]interface{}{
		"test_map": map[interface{}]interface{}{
			"key1":                 "value",
			&stringGeneric{"key2"}: 2,
			"key3": map[interface{}]interface{}{
				"subkey1": []interface{}{"subvalue"},
			},
		},
	}
	expectMap := map[string]interface{}{
		"test_map": map[string]interface{}{
			"key1": "value",
			"key2": 2,
			"key3": map[string]interface{}{
				"subkey1": []interface{}{"subvalue"},
			},
		},
	}

	d, err := mapToJsonable(testMap)
	expect(t, expectMap, d)
	expect(t, nil, err)
}

func TestMapInputSource_sliceToJsonable(t *testing.T) {
	testSlice := []interface{}{
		map[interface{}]interface{}{
			"key1":                 "value",
			&stringGeneric{"key2"}: 2,
			"key3": map[interface{}]interface{}{
				"subkey1": []interface{}{"subvalue"},
			},
		},
		"value",
		2,
	}
	expectSlice := []interface{}{
		map[string]interface{}{
			"key1": "value",
			"key2": 2,
			"key3": map[string]interface{}{
				"subkey1": []interface{}{"subvalue"},
			},
		},
		"value",
		2,
	}

	d, err := sliceToJsonable(testSlice)
	expect(t, expectSlice, d)
	expect(t, nil, err)
}

func TestMapInputSource_Json(t *testing.T) {
	inputSource := NewMapInputSource(
		"test",
		map[interface{}]interface{}{
			"test_map": map[interface{}]interface{}{
				"key1":                 "value",
				&stringGeneric{"key2"}: 2,
				"key3": []interface{}{map[interface{}]interface{}{
					"subkey1": []interface{}{"subvalue"},
				}},
			},
			"test_slice": []interface{}{
				"value",
				2,
				map[interface{}]interface{}{
					"key1":                 "value",
					&stringGeneric{"key2"}: 2,
					"key3": []interface{}{map[interface{}]interface{}{
						"subkey1": []interface{}{"subvalue"},
					}},
					"key4": []interface{}{
						[]interface{}{1, 4, 7},
						[]interface{}{2, 5, 8},
						[]interface{}{3, 6, 9},
					},
				},
			},
		})

	d, err := inputSource.Json("test_map")
	expect(t, []byte(`{"key1":"value","key2":2,"key3":[{"subkey1":["subvalue"]}]}`), d)
	expect(t, nil, err)

	d, err = inputSource.Json("test_map.key3")
	expect(t, []byte(`[{"subkey1":["subvalue"]}]`), d)
	expect(t, nil, err)

	d, err = inputSource.Json("test_slice")
	expect(t, []byte(`["value",2,{"key1":"value","key2":2,"key3":[{"subkey1":["subvalue"]}],"key4":[[1,4,7],[2,5,8],[3,6,9]]}]`), d)
	expect(t, nil, err)
}
