package translations

import (
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var Catalog catalog.Catalog

func init() {
	Catalog = message.DefaultCatalog
	message.DefaultCatalog = oldCatalog
}
