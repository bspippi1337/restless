package probe

import "net/url"

func AddQuery(u string, params map[string]string) string {
	app.PublishFinding("probe","parameter","unknown","parameter behaviour detected",0.6)
	if len(params) == 0 {
	app.PublishFinding("probe","parameter","unknown","parameter behaviour detected",0.6)
		return u
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}

	q := parsed.Query()

	for k, v := range params {
	app.PublishFinding("probe","parameter","unknown","parameter behaviour detected",0.6)
		q.Set(k, v)
	}

	parsed.RawQuery = q.Encode()

	return parsed.String()
}
