package or

import "sync"

func Or(channels ...<-chan interface{}) <-chan interface{} {
	active := make([]<-chan interface{}, 0, len(channels))
	for _, ch := range channels {
		if ch != nil {
			active = append(active, ch)
		}
	}

	switch len(active) {
	case 0:
		return nil
	case 1:
		return active[0]
	}

	orDone := make(chan interface{})
	var once sync.Once

	closeOrDone := func() {
		once.Do(func() {
			close(orDone)
		})
	}

	for _, ch := range active {
		go func(ch <-chan interface{}) {
			for {
				select {
				case _, ok := <-ch:
					if !ok {
						closeOrDone()
						return
					}
				case <-orDone:
					return
				}
			}
		}(ch)
	}

	return orDone
}
