package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	// "syscall"
	"strings"
	// "io/ioutil"
	"net/url"
	// "bytes"
	// dmp "github.com/sergi/go-diff/diffmatchpatch"
	humioapi "github.com/humio/cli/api"
	"github.com/shurcooL/graphql"
)

type DashboardsData struct {
	Dashboards []struct {
		DisplayName string
		TemplateYaml string
	} `graphql:... on View`
}

func main() {
	userAgent := fmt.Sprintf("humio-go-client/cloud-tools")
	humioAddress := os.Getenv("HUMIO_ADDRESS")
	humioToken := os.Getenv("HUMIO_TOKEN")
	currView := os.Getenv("HUMIO_VIEW")
	// tmpDir := os.Getenv("TMPDIR")

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
	views := client.Views()
	view, err := views.Get(currView)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Working on View:", view.Name)

	// Get all the dashboards created in the view
	// Save all view's dashboards to files
	result := get_view_dashboards(currView, client)

	for i := range result.Dashboards {
		dashboardName := result.Dashboards[i].DisplayName
		dashboardTmpl := result.Dashboards[i].TemplateYaml
		write_to_file("tmp_" + dashboardName, dashboardTmpl)
		log.Println(dashboardName)
		diffValue , _ := diff_template_files("tmp_" + dashboardName + ".yaml", strings.ReplaceAll(dashboardName, " ", "") + ".yaml")
		if len(diffValue) > 0 { log.Println(string(diffValue)) }
	}

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

func write_to_file(fileName string, content string) {
	v := []byte(content)
	f, err := os.Create(fileName + ".yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(v)
	f.Sync()
}

func diff_template_files(origin, target string) ([]byte, bool) {
	// originFile, err := ioutil.ReadFile(origin)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// targetFile, err := ioutil.ReadFile(target)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// dmp := dmp.New()
	// diffs := dmp.DiffMain(string(originFile), string(targetFile), true)
	// log.Println(dmp.DiffPrettyText(diffs))
	
	cmd := exec.Command("diff", origin, target)
	cmdOutput, _ := cmd.Output()
	// if err != nil {
	// 	log.Println("has err:", err)
	// 	return nil, false
	// }
	if cmdOutput != nil {
		return cmdOutput, true
	}
	return nil, false

	// return bytes.Equal(targetFile, originFile)
}
