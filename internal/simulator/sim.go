package simulator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Run(defaultURL string) (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Method [GET]: ")
	m, _ := reader.ReadString(10) // 10 == newline
	method := strings.TrimSpace(m)
	if method == "" {
		method = "GET"
	}

	fmt.Print("URL [" + defaultURL + "]: ")
	u, _ := reader.ReadString(10)
	url := strings.TrimSpace(u)
	if url == "" {
		url = defaultURL
	}

	return method, url
}
