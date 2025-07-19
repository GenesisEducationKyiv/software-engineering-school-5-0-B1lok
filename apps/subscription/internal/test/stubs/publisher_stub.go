package stubs

import "context"

type PublisherStub struct {
	PublishFn func(ctx context.Context, queue string, payload []byte) error
}

func NewPublisherStub() *PublisherStub {
	return &PublisherStub{
		PublishFn: nil,
	}
}

func (s *PublisherStub) Publish(ctx context.Context, queue string, payload []byte) error {
	if s.PublishFn != nil {
		return s.PublishFn(ctx, queue, payload)
	}
	return nil
}
