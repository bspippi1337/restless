package graph
import(
"strings"
"github.com/bspippi1337/restless/internal/store"
)
func Render(api store.API)string{
var b strings.Builder
b.WriteString("API\n")
for _,e:=range api.Endpoints{
b.WriteString("├ ")
b.WriteString(e.Path)
b.WriteString(" ")
b.WriteString(strings.Join(e.Methods,","))
b.WriteString("\n")
}
return b.String()
}
