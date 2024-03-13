package alpha2

import (
	"github.com/meschbach/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
)

const Version = "alpha2"

var Type = resources.Type{
	Kind:    service.Kind,
	Version: Version,
}
