package listings

// BeforeBrowse reports whether a belongs before b in the requested public
// ordering. UUID is the deterministic final tie-breaker for every sort.
func BeforeBrowse(a, b Listing, sort string) bool {
	switch sort {
	case "price_asc":
		if *a.PriceCentavos != *b.PriceCentavos {
			return *a.PriceCentavos < *b.PriceCentavos
		}
		return a.ID.String() < b.ID.String()
	case "price_desc":
		if *a.PriceCentavos != *b.PriceCentavos {
			return *a.PriceCentavos > *b.PriceCentavos
		}
		return a.ID.String() > b.ID.String()
	default:
		if !a.PublishedAt.Equal(*b.PublishedAt) {
			return a.PublishedAt.After(*b.PublishedAt)
		}
		return a.ID.String() > b.ID.String()
	}
}

// AfterCursor reports whether a listing belongs on the page after cursor.
func AfterCursor(l Listing, cursor *ListCursor, sort string) bool {
	if cursor == nil {
		return true
	}
	switch sort {
	case "price_asc":
		return *l.PriceCentavos > cursor.PriceCentavos || (*l.PriceCentavos == cursor.PriceCentavos && l.ID.String() > cursor.ID.String())
	case "price_desc":
		return *l.PriceCentavos < cursor.PriceCentavos || (*l.PriceCentavos == cursor.PriceCentavos && l.ID.String() < cursor.ID.String())
	default:
		return l.PublishedAt.Before(cursor.PublishedAt) || (l.PublishedAt.Equal(cursor.PublishedAt) && l.ID.String() < cursor.ID.String())
	}
}
