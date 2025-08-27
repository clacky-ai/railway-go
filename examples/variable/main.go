package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/railwayapp/cli/pkg/railway"
)

func main() {
	// 检查环境变量
	apiToken := os.Getenv("RAILWAY_API_TOKEN")
	if apiToken == "" {
		log.Fatal("请设置 RAILWAY_API_TOKEN 环境变量")
	}

	// 创建 Railway 客户端
	cli, err := railway.New(
		railway.WithAPIToken(apiToken),
		railway.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	ctx := context.Background()

	project, err := cli.GetProject(ctx, "538cdb06-4ac5-4a62-8166-5a2b0c48e4b6")
	check(err)

	serviceID := "b9c1597b-258b-4426-90c2-3dc337309e29"

	variables, err := cli.GetVariables(ctx, project.ID, project.Environments[0].ID, serviceID)
	check(err)
	fmt.Println(variables)

	config, err := cli.GetEnvironmentConfig(ctx, project.Environments[0].ID, true, true)
	check(err)
	fmt.Println(config)

	fmt.Println(config.EnvironmentStagedChanges.Patch)
	//services := config.EnvironmentStagedChanges.Patch["services"]
	marshal, _ := json.Marshal(config.EnvironmentStagedChanges.Patch)
	// {"services":{"b9c1597b-258b-4426-90c2-3dc337309e29":{"variables":{"ssssssssssssss":null}}}}

	fmt.Println(string(marshal))

	if servicesItem := config.EnvironmentStagedChanges.Patch["services"]; servicesItem != nil {
		if services, ok := servicesItem.(map[string]interface{}); ok {
			if serviceItem := services[serviceID]; serviceItem != nil {
				if service, ok := serviceItem.(map[string]interface{}); ok {
					if variables := service["variables"]; variables != nil {
						if variables, ok := variables.(map[string]interface{}); ok {
							for k, v := range variables {
								if v == nil {
									fmt.Println("delete key:" + k)
								} else {
									obj := v.(map[string]interface{})
									value := obj["value"]
									var env string
									switch value.(type) {
									case string:
										env = value.(string)
									default:
										env = fmt.Sprintf("%v", v)
									}
									fmt.Println(k, env)
								}

							}
						}
					}
				}
			}
		}
	}
	//if services != nil {
	//	m, ok := services[serviceID](map[string]interface{})
	//	if ok {
	//		variables := m["variables"]
	//		m, ok := variables.(map[string]*string)
	//	}
	//	app := services[serviceID]
	//
	//}
	//for k, v := range services[serviceID]["variables"].(map[string]interface{}) {
	//	fmt.Println(k, v)
	//}
	//services[serviceID]["variables"] = m

	v1 := "vss"
	m := map[string]*string{}
	m["kss"] = &v1

	s, err := cli.StageServiceVariables(ctx, project.Environments[0].ID, serviceID, m)
	check(err)
	println(s)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
