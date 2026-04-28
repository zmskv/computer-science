package or

import (
	"testing"
	"time"
)

func TestOrNoChannelsReturnsNil(t *testing.T) {
	if got := Or(); got != nil {
		t.Fatal("Or() = non-nil, want nil")
	}
}

func TestOrAllNilChannelsReturnsNil(t *testing.T) {
	var ch1 <-chan interface{}
	var ch2 <-chan interface{}

	if got := Or(ch1, ch2); got != nil {
		t.Fatal("Or(nil, nil) = non-nil, want nil")
	}
}

func TestOrSingleChannelReturnsOriginalChannel(t *testing.T) {
	ch := make(chan interface{})

	if got := Or(ch); got != ch {
		t.Fatal("Or(ch) did not return the original channel")
	}
}

func TestOrClosesWhenAnyChannelCloses(t *testing.T) {
	slow := make(chan interface{})
	fast := make(chan interface{})

	done := Or(slow, fast)

	select {
	case <-done:
		t.Fatal("or-channel closed before any input channel was closed")
	default:
	}

	close(fast)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("or-channel did not close after one input channel closed")
	}
}

func TestOrClosesIfChannelIsAlreadyClosed(t *testing.T) {
	closed := make(chan interface{})
	close(closed)

	open := make(chan interface{})
	done := Or(open, closed)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("or-channel did not close when one of the channels was already closed")
	}
}

func TestOrIgnoresNilChannels(t *testing.T) {
	var nilCh <-chan interface{}
	slow := make(chan interface{})
	fast := make(chan interface{})

	done := Or(nilCh, slow, fast)
	close(fast)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("or-channel did not close when a non-nil input channel closed")
	}
}

func TestOrDoesNotCloseOnValueSend(t *testing.T) {
	valueCh := make(chan interface{}, 1)
	blocker := make(chan interface{})

	done := Or(valueCh, blocker)
	valueCh <- struct{}{}

	select {
	case <-done:
		t.Fatal("or-channel closed after receiving a value instead of waiting for channel closure")
	case <-time.After(20 * time.Millisecond):
	}

	close(blocker)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("or-channel did not close after blocker channel closed")
	}
}
