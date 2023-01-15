package brev

type channel struct {
	name string

	events      list[message]
	subscribers []subscriber
}

type message struct {
	payload []byte
}

type subscriber struct {
	callback string
}

func newChannel(name string) *channel {
	c := &channel{
		name:        name,
		events:      list[message]{},
		subscribers: make([]subscriber, 0),
	}

	return c
}
