package teacher

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func runCapture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return out.String(), err
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func must(err error, msg string) {
	if err != nil {
		fmt.Println("FAIL:", msg)
		os.Exit(1)
	}
}

func Run() {
	fmt.Println("Restless Teacher (CI mode)")
	fmt.Println("---------------------------")

	must(run("go", "version"), "Go not available")
	must(run("go", "build", "-o", "restless", "./cmd/restless"), "Build failed")

	out, err := runCapture("./restless", "openapi", "import",
		"https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/examples/v3.0/petstore.json")
	must(err, "Import failed")

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		fmt.Println("FAIL: could not parse spec ID")
		os.Exit(1)
	}

	id := strings.TrimPrefix(lines[len(lines)-1], "imported: ")
	id = strings.TrimSpace(id)

	if id == "" {
		fmt.Println("FAIL: empty spec ID")
		os.Exit(1)
	}

	fmt.Println("Spec ID:", id)

	must(run("./restless", "openapi", "ls"), "Listing failed")

	os.Setenv("RESTLESS_STRICT", "1")

	fmt.Println("Running endpoint explicitly with base...")

	must(run("./restless", "openapi", "run",
		id,
		"GET",
		"/pets",
		"--base",
		"https://petstore3.swagger.io/api/v3"),
		"Endpoint run failed")

	fmt.Println()
	fmt.Println("Teacher completed successfully.")
}
