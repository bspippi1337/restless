package probe

import "net/url"

func AddQuery(u string, params map[string]string) string {
	if len(params) == 0 {
		return u
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	q := parsed.Query()

	for k, v := range params {
		q.Set(k, v)
	}

	parsed.RawQuery = q.Encode()

	return parsed.String()
}
