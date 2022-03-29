package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	// "strings"
	"net/url"
	"path/filepath"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	humioapi "github.com/humio/cli/api"
	"github.com/shurcooL/graphql"
)

type DashboardsData struct {
	Dashboards []struct {
		DisplayName string
		TemplateYaml string
	} `graphql:... on View`
}

type DashboardMeta struct {
	Name string `yaml:"name"`
}

func main() {
	userAgent := fmt.Sprintf("humio-go-client/cloud-tools")
	humioAddress := os.Getenv("HUMIO_ADDRESS")
	humioToken := os.Getenv("HUMIO_TOKEN")
	currView := os.Getenv("HUMIO_VIEW")
	dashboardDir := "."

	addr, err := url.Parse(humioAddress)
	if err != nil {
		log.Println(err)
	}

	// Create a config for Humio Client
	clientConfig := humioapi.Config{
		UserAgent: userAgent,
		Address: addr,
		Token: humioToken,
	}
	client := humioapi.NewClient(clientConfig)

	// Try to get the target view
	viewsClient := client.Views()
	view, err := viewsClient.Get(currView)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Working on View:", view.Name)

	// List all the dashboard files from local directory
	localDashboards := get_local_dashboards(dashboardDir)

	for i := range localDashboards {
		fileContent, err := ioutil.ReadFile(localDashboards[i])
		if err != nil {
			log.Println(err)
		}

		var dashboardName DashboardMeta
		err = yaml.Unmarshal(fileContent, &dashboardName)
		if err != nil {
			log.Fatal(err)
		}
		err = create_dashboard_from_file(currView, dashboardName.Name, string(fileContent), client)
		if err != nil {
			log.Println(err)
		}
	}

	// // Get all the dashboards created in the view
	// // 
	// result := get_view_dashboards(currView, client)

	// // Create a temp dir to store downloaded dashboards
	// tempDir, _ := os.MkdirTemp(".", ".cache")

	// // Save all view's dashboards to files
	// for i := range result.Dashboards {
	// 	dashboardName := result.Dashboards[i].DisplayName
	// 	dashboardTmpl := result.Dashboards[i].TemplateYaml
	// 	write_to_file(tempDir, dashboardName, dashboardTmpl)

	// 	// diffValue := diff_template_files("tmp_" + dashboardName + ".yaml", dashboardName + ".yaml")
	// 	// if len(diffValue) > 0 { log.Println(string(diffValue)) }
	// }

}

func get_local_dashboards(dirPath string) []string {
	files := make([]string, 0)

	items, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		return nil
	}
	for i := range items {
		if !items[i].IsDir() && filepath.Ext(items[i].Name()) == ".yaml" {
			files = append(files, items[i].Name())
		}
	}
	return files
}

func create_dashboard_from_file(viewName, dashboardName, template string, client *humioapi.Client) error {
	var mutation struct {
		CreateDashboardFromTemplate struct {
			Typename graphql.String `graphql:"__typename"`
		} `graphql:"createDashboardFromTemplate(input: {searchDomainName: $viewname, overrideName: $dashboardname, template: $template})"`
	}

	variables := map[string]interface{}{
		"viewname": graphql.String(viewName),
		"dashboardname": graphql.String(dashboardName),
		"template": graphql.String(template),
	}

	return client.Mutate(&mutation, variables)
}

func get_view_dashboards(viewName string, client *humioapi.Client) DashboardsData {
	var query struct {
		Result DashboardsData `graphql:"searchDomain(name: $name)"`
	}

	variables := map[string]interface{}{
		"name": graphql.String(viewName),
	}

	err := client.Query(&query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Result
}

func write_to_file(dirName, fileName, content string) {
	filePath := filepath.Join(dirName, "/", fileName + ".yaml")
	v := []byte(content)
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(v)
}

func diff_template_files(origin, target string) []byte {	
	cmd := exec.Command("diff", origin, target)
	cmdOutput, _ := cmd.Output()
	// if err != nil {
	// 	log.Debug(err)
	// }
	if cmdOutput != nil {
		return cmdOutput
	}
	return nil
}
