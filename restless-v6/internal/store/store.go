package store
type Endpoint struct{
Path string
Methods []string
}
type API struct{
Base string
Endpoints []Endpoint
}
var last API
func Save(a API){last=a}
func Last()API{return last}
