package async

type IterableMap[keyT comparable, valT any] struct {
	Map[keyT, valT]
}

func NewIterableMap[keyT comparable, valT any](capacity int) *IterableMap[keyT, valT] {
	return &IterableMap[keyT, valT]{
		Map: *NewMap[keyT, valT](capacity),
	}
}

func (i *IterableMap[keyT, valT]) StartIteration() (<-chan valT, chan<- struct{}) {
	chIteration := make(chan valT)
	chStopIteration := make(chan struct{})

	go func() {
		i.rwMu.RLock()

		for _, val := range i.storage {
			i.rwMu.RUnlock()
			select {
			case <-chStopIteration:
				close(chStopIteration)
				close(chIteration)

				return
			default:
				chIteration <- val

				i.rwMu.RLock()
			}
		}

		close(chIteration)
		close(chStopIteration)
		i.rwMu.RUnlock()
	}()

	return chIteration, chStopIteration
}
