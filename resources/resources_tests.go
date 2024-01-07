package resources

import "github.com/go-faker/faker/v4"

func FakeMeta(opts ...fakeOpt) Meta {
	return FakeMetaOf(FakeType(), opts...)
}

type fakeMetaConfig struct {
	nameGenerator func() string
}
type fakeOpt func(config *fakeMetaConfig)

func newFakeMetas(opts []fakeOpt) *fakeMetaConfig {
	f := &fakeMetaConfig{
		nameGenerator: func() string {
			return faker.Word()
		},
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func WithNamingDomain(fn func() string) fakeOpt {
	return func(config *fakeMetaConfig) {
		config.nameGenerator = fn
	}
}

func FakeMetaOf(someType Type, opts ...fakeOpt) Meta {
	cfg := newFakeMetas(opts)
	return Meta{
		Type: someType,
		Name: cfg.nameGenerator(),
	}
}

func FakeType() Type {
	return Type{
		Kind:    faker.DomainName(),
		Version: faker.Word(),
	}
}
