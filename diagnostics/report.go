package diagnostics

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/elastic/go-sysinfo"
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
	SchemaJSON        string
	Modules           []string
	ModulesJSON       string
	ProfileImg        string
	HostJSON          string
	PrometheusMetrics string
	Validations       []Validation
}

var globalConfig Config

func generateClient(clientUrl string, authEnabled bool) weaviate.Config {

	var config weaviate.Config

	parsedURL, err := url.Parse(clientUrl)
	if err != nil {
		log.Fatal("Cannot parse Weaviate url:", err)
	}

	if !authEnabled {
		config = weaviate.Config{
			Scheme: parsedURL.Scheme,
			Host:   parsedURL.Host,
		}
		return config
	}

	username := getInput("Username:", ' ')
	password := getInput("Password:", '*')

	authConfig, err := weaviate.NewConfig(
		parsedURL.Host,
		parsedURL.Scheme,
		auth.ResourceOwnerPasswordFlow{
			Username: username,
			Password: password,
		},
		nil)
	if err != nil {
		log.Fatal("Cannot create authenticated Weaviate config:", err)
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

	// Print a banner for Weaviate in ascii art
	fmt.Println(`  _ _ _ 
  | | | | ` + green("Weaviate Diagnostics") + `
  |__/_/ 	   
`)

	fmt.Printf("- Retrieving Weaviate schema from: %s\n", cyan(globalConfig.Url))
	fmt.Printf("- Authentication: %s\n", cyan(globalConfig.Auth))

	config := generateClient(globalConfig.Url, globalConfig.Auth)

	client := weaviate.New(config)
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

		parsed, err := json.Marshal(node)
		if err != nil {
			log.Fatal("Cannot parse Weaviate node info:", err)
		}

		nodesJSON = append(nodesJSON, parsed...)

	}

	var prometheusMetrics []byte = []byte{}
	resp, err := http.Get(globalConfig.MetricsUrl)
	if err != nil {
		fmt.Printf("%s Skipping prometheus metrics\n", red("x"))
	} else {
		prometheusMetrics, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Cannot parse Weaviate prometheus metrics:", err)
		}
		fmt.Printf("%s Prometheus metrics retrieved\n", green("✓"))
	}
	defer resp.Body.Close()

	hostData, err := sysinfo.Host()

	if err != nil {
		panic(err)
	}

	memory, err := hostData.Memory()

	if err != nil {
		panic(err)
	}
	hostJSON, err := json.Marshal(memory)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s Host data retrieved\n", green("✓"))

	tmplt, err := template.ParseFiles("diagnostics/templates/report.html")
	if err != nil {
		log.Fatal("Cannot parse report template:", err)
	}

	fmt.Printf("- Generating CPU profile..\n")
	profile := getProf(globalConfig.ProfileUrl)
	fmt.Printf("%s CPU profile retrieved\n", green("✓"))

	validations := validate(schema)

	report := Report{
		Meta:              meta,
		Date:              time.Now().Format(time.RFC3339),
		Nodes:             nodes.Nodes,
		NodesJSON:         string(nodesJSON),
		MetaJSON:          string(metaJSON),
		SchemaJSON:        string(schemaJSON),
		Modules:           moduleList,
		ModulesJSON:       string(modulesJSON),
		ProfileImg:        profile,
		HostJSON:          string(hostJSON),
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

	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s Report written to %s\n\n", green("✓"), yellow(globalConfig.OutputFile))

}
