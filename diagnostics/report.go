package diagnostics

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate/entities/models"
)

type Report struct {
	Meta              *models.Meta
	Date              string
	Nodes             []*models.NodeStatus
	NodesJSON         string
	MetaJSON          string
	TotalClasses      int
	SchemaJSON        string
	Modules           []string
	ModulesJSON       string
	ProfileImg        string
	HostInformation   HostInfo
	PrometheusMetrics string
	Validations       []Validation
}

var globalConfig Config

//go:embed templates/report.html
var templateFile []byte

func generateClient(clientUrl string, authMethod string) weaviate.Client {

	var config weaviate.Config

	parsedURL, err := url.Parse(clientUrl)

	if err != nil {
		log.Fatal("Cannot parse Weaviate url:", err)
	}

	if authMethod == "none" {
		config = weaviate.Config{
			Scheme: parsedURL.Scheme,
			Host:   parsedURL.Host,
		}
	}

	if authMethod == "apiKey" {
		config = weaviate.Config{
			Scheme:     parsedURL.Scheme,
			Host:       parsedURL.Host,
			AuthConfig: auth.ApiKey{Value: globalConfig.ApiKey},
		}
	}

	if authMethod == "oidc" {

		username := globalConfig.User
		password := globalConfig.Pass

		if username == "" {
			username = getInput("Username:", ' ')
		}
		if password == "" {
			password = getInput("Password:", '*')
		}

		config = weaviate.Config{
			Scheme: parsedURL.Scheme,
			Host:   parsedURL.Host,
			AuthConfig: auth.ResourceOwnerPasswordFlow{
				Username: username,
				Password: password,
			},
		}
	}

	authConfig, err := weaviate.NewClient(config)

	if err != nil {
		log.Fatal("Cannot create Weaviate config:", err)
	}

	return *authConfig
}

func getInput(label string, mask rune) string {
	prompt := promptui.Prompt{}

	templates := &promptui.PromptTemplates{
		Prompt:  "- {{ . }}",
		Valid:   "- {{ . }} ",
		Success: "- {{ . }} ",
	}

	if mask == ' ' {
		prompt = promptui.Prompt{
			Label:     label,
			Templates: templates,
		}
	} else {
		prompt = promptui.Prompt{
			Label:     label,
			Mask:      mask,
			Templates: templates,
		}
	}

	input, err := prompt.Run()
	if err != nil {
		log.Fatal("Cannot parse prompt:", err)
	}
	return input
}

func GenerateReport() {

	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Print a banner for Weaviate in ascii art
	fmt.Println(`  _ _ _ 
  | | | | ` + green("Weaviate Diagnostics") + `
  |__/_/ 	   
`)

	fmt.Printf("- Retrieving Weaviate schema from: %s\n", cyan(globalConfig.Url))

	authMethod := "none"
	if globalConfig.User != "" {
		authMethod = "oidc"
	}
	if globalConfig.ApiKey != "" {
		authMethod = "apiKey"
	}

	fmt.Printf("- Authentication: %s\n", cyan(authMethod))

	client := generateClient(globalConfig.Url, authMethod)

	metaGetter := client.Misc().MetaGetter()
	meta, err := metaGetter.Do(context.Background())
	if err != nil {
		log.Fatal("Cannot retrieve Weaviate /v1/meta", err)
	}

	fmt.Printf("%s Meta retrieved\n", green("✓"))

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		log.Fatal("Cannot parse Weaviate /v1/meta:", err)
	}

	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		log.Fatal("Cannot retrieve Weaviate /v1/schema:", err)
	}

	fmt.Printf("%s Schema retrieved\n", green("✓"))

	modules, ok := meta.Modules.(map[string]interface{})
	if !ok {
		log.Fatal("Cannot parse Weaviate schema:", err)
	}

	moduleList := []string{}
	for k := range modules {
		moduleList = append(moduleList, k)
	}

	modulesJSON, err := json.Marshal(meta.Modules)
	if err != nil {
		log.Fatal("Cannot parse Weaviate modules:", err)
	}

	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatal("Cannot parse Weaviate schema:", err)
	}

	nodes, err := client.Cluster().NodesStatusGetter().Do(context.Background())
	if err != nil {
		log.Fatal("Cannot retrieve Weaviate /v1/nodes:", err)
	}

	fmt.Printf("%s Nodes status retrieved\n", green("✓"))

	nodesJSON := []byte{}
	for _, node := range nodes.Nodes {

		parsed, err := json.MarshalIndent(node, "", "  ")
		if err != nil {
			log.Fatal("Cannot parse Weaviate node info:", err)
		}

		nodesJSON = append(nodesJSON, parsed...)

	}

	var prometheusMetrics []byte = []byte{}
	resp, err := http.Get(globalConfig.MetricsUrl)
	if err != nil {
		fmt.Printf("%s Skipping prometheus metrics: %s\n", red("x"), err)
	} else {
		prometheusMetrics, err = io.ReadAll(resp.Body)
		// limit the amount of metrics to 100k bytes
		if len(prometheusMetrics) > 500000 {
			prometheusMetrics = prometheusMetrics[:500000]
			prometheusMetrics = append(prometheusMetrics, []byte(".. truncated due to size")...)
		}
		if err != nil {
			log.Fatal("Cannot parse Weaviate prometheus metrics:", err)
		}
		fmt.Printf("%s Prometheus metrics retrieved\n", green("✓"))
		defer resp.Body.Close()
	}

	hostInformation := getHostInfo()
	fmt.Printf("%s Host data retrieved\n", green("✓"))

	tmplt := template.Must(template.New("report").Parse(string(templateFile)))

	fmt.Printf("- Generating CPU profile..\n")
	profile := getProf(globalConfig.ProfileUrl)
	fmt.Printf("%s CPU profile retrieved\n", green("✓"))

	validations := validate(schema)
	fmt.Printf("%s Running validation checks\n", green("✓"))

	report := Report{
		Meta:              meta,
		Date:              time.Now().Format(time.RFC3339),
		Nodes:             nodes.Nodes,
		NodesJSON:         string(nodesJSON),
		MetaJSON:          string(metaJSON),
		TotalClasses:      len(schema.Classes),
		SchemaJSON:        string(schemaJSON),
		Modules:           moduleList,
		ModulesJSON:       string(modulesJSON),
		ProfileImg:        profile,
		HostInformation:   hostInformation,
		PrometheusMetrics: string(prometheusMetrics),
		Validations:       validations,
	}

	outputFile, err := os.Create(globalConfig.OutputFile)
	if err != nil {
		log.Fatal("Cannot create report file:", err)
	}
	err = tmplt.Execute(outputFile, report)
	if err != nil {
		log.Fatal("Cannot write report file:", err)
	}
	fmt.Printf("%s Report written to %s\n\n", green("✓"), yellow(globalConfig.OutputFile))
}
