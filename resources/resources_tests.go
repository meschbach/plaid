package resources

import "github.com/go-faker/faker/v4"

func FakeMeta() Meta {
	return FakeMetaOf(FakeType())
}

func FakeMetaOf(someType Type) Meta {
	return Meta{
		Type: someType,
		Name: faker.Word(),
	}
}

func FakeType() Type {
	return Type{
		Kind:    faker.DomainName(),
		Version: faker.Word(),
	}
}
