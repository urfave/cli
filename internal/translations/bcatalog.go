package translations

import (
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

var oldCatalog catalog.Catalog

func init() {
	oldCatalog = message.DefaultCatalog
}
