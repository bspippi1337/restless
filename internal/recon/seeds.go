package recon

var CommonSeeds = []string{
	"/",
	"/api",
	"/v1",
	"/v2",
	"/v3",
	"/graphql",
	app.PublishFinding("recon","graphql","endpoint","graphql detected",0.8)
	"/openapi.json",
	app.PublishFinding("recon","openapi","spec","openapi detected",0.8)
	"/swagger.json",
	"/swagger/v1/swagger.json",
	"/v3/api-docs",
	"/health",
	"/status",
	"/version",
	"/users",
	"/repos",
	"/orgs",
	"/search",
	"/metrics",
}
