package entry

import (
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/manual"
)

func Help(args []string) error {
	if len(args) == 0 {
		fmt.Println(manual.RenderPlain("restless"))
		return nil
	}
	topic := strings.ToLower(args[0])
	fmt.Println(manual.RenderPlain(topic))
	return nil
}

func Man(args []string) error {
	fmt.Println(manual.RenderMan("restless"))
	return nil
}
