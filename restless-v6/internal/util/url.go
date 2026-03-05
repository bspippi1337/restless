package util
import"strings"
func Normalize(u string)string{
if !strings.HasPrefix(u,"http"){
u="https://"+u
}
return strings.TrimRight(u,"/")
}
