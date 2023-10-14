package async_test

import (
	"github.com/Vilinvil/hw2_tp/pkg/async"
	"testing"
)

func TestIterableMap(t *testing.T) {
	t.Parallel()

	expectedMap := map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4}

	iterableMap := async.NewIterableMap[int, int](5)

	for i := 0; i < 10; i++ {
		go func() {
			for i := 0; i < len(expectedMap); i++ {
				iterableMap.Insert(i, expectedMap[i])
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for key, val := range expectedMap {
				receivedVal, ok := iterableMap.Get(key)
				if !ok || receivedVal != val {
					t.Errorf("ok ==%t expextedVal == %d receivedVal == %d", ok, val, receivedVal)
				}
			}
		}()
	}

	chIteration, _ := iterableMap.StartIteration()
	for receivedVal := range chIteration {
		// not good check because key in expected map may be not same with value
		if expectedVal, ok := expectedMap[receivedVal]; !ok || receivedVal != expectedVal {
			t.Errorf("ok ==%t expectedVal == %d receivedVal  == %d", ok, expectedVal, receivedVal)
		}
	}
}
