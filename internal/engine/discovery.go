package engine

func DiscoverEndpoints(target string) []string {

	return []string{
		"/users",
		"/repos",
		"/issues",
		"/graphql",
	}
}
