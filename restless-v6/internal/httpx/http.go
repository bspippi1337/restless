package httpx
import(
"net/http"
"time"
)
func Client()*http.Client{
return &http.Client{Timeout:10*time.Second}
}
