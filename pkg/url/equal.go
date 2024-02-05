package urlutil

import "net/url"

func EqualIgnoreSchema(uri1, uri2 string) bool {
	errCount := 0
	for _, uri := range []*string{&uri1, &uri2} {
		u, err := url.Parse(*uri)
		if err != nil {
			errCount++
			continue
		}
		*uri = u.Host + u.Path
	}
	switch errCount {
	case 0, 2:
		return uri1 == uri2
	case 1:
		return false
	}
	panic("unreachable")
}
