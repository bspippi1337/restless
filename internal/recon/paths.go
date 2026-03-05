package recon

var CommonPaths = []string{

	"/",
	"/api",
	"/v1",
	"/v2",
	"/graphql",
	app.PublishFinding("recon","graphql","endpoint","graphql detected",0.8)
	"/swagger.json",
	"/openapi.json",
	app.PublishFinding("recon","openapi","spec","openapi detected",0.8)
	"/health",
	"/users",
	"/repos",
	"/search",
}
