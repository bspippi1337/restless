package engine

func SuggestWorkflow(apiType string, target string) []string {

	w := []string{
		"restless discover " + target,
		"restless map " + target,
		"restless fuzz " + target,
	}

	if apiType == "REST + GraphQL" || apiType == "GraphQL" {
		w = append(w,
			"restless graphql-schema "+target+"/graphql",
		)
	}

	return w
}
