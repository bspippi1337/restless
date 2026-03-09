package engine

func mergeTemplates(target string, endpoints *[]Endpoint, seen map[string]bool) {

	templates := ExtractTemplatesFromRoot(target)

	for _, p := range templates {

		p = normalizeTemplate(p)

		if p == "" || p == "/" {
			continue
		}

		if !seen[p] {

			seen[p] = true

			*endpoints = append(*endpoints, Endpoint{
				Path:       p,
				Confidence: "high",
			})

		}
	}
}
