package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/discovery"
	"github.com/spf13/cobra"
)

func NewInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <target>",
		Short: "Fingerprint and inspect a target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			fmt.Println("Restless API Discovery Engine")
			fmt.Println()

			fp, err := discovery.FingerprintTarget(target)
			if err != nil {
				return err
			}

			fmt.Printf("Scanning: %s\n\n", fp.Target)

			fmt.Println("[1/5] probing target")
			fmt.Println("[2/5] fingerprinting stack")
			fmt.Println("[3/5] extracting API hints")
			fmt.Println("[4/5] inferring architecture")
			fmt.Println("[5/5] scoring confidence")

			fmt.Println()
			fmt.Println("Fingerprint")
			fmt.Println("-----------")
			fmt.Printf("API type: %s\n", fp.APIType)
			fmt.Printf("Server: %s\n", fp.Server)
			fmt.Printf("Confidence: %d/100\n", fp.Confidence)

			fmt.Println()
			fmt.Println("Technologies")
			fmt.Println("------------")

			if len(fp.Technologies) == 0 {
				fmt.Println("No obvious technologies detected")
			} else {
				for _, t := range fp.Technologies {
					fmt.Printf("- %s\n", t)
				}
			}

			fmt.Println()
			fmt.Println("Discovery")
			fmt.Println("---------")

			if fp.GraphQL {
				fmt.Println("- GraphQL hints detected")
			}

			if fp.OpenAPI {
				fmt.Println("- OpenAPI/Swagger hints detected")
			}

			if len(fp.InterestingURLs) == 0 {
				fmt.Println("- No obvious API endpoints discovered")
			} else {
				for _, u := range fp.InterestingURLs {
					fmt.Printf("- %s\n", u)
				}
			}

			return nil
		},
	}

	return cmd
}
