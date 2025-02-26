package azure_devops_api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/work"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

type Configuration struct {
	PersonalAccessToken string `json:"PersonalAccessToken"`
	OrganizationName    string `json:"OrganizationName"`
	TeamName            string `json:"TeamName"`
	ProjectName            string `json:"ProjectName"`
	SprintName            string `json:"SprintName"`
}

type Wobject struct {
	Id string
}

func NotMain() {
    coreClient, ctx, err := GetCoreClientAndCtx(Configuration{})
	if err != nil {
		log.Fatal(err)
	}
	// Get first page of the list of team projects for your organization
	responseValue, err := coreClient.GetProjects(ctx, core.GetProjectsArgs{})
	if err != nil {
		log.Fatal(err)
	}

	index := 0
	for responseValue != nil {
		// Log the page of team project names
		for _, teamProjectReference := range (*responseValue).Value {
			log.Printf("Name[%v] = %v\n", index, *teamProjectReference.Name)
			index++
		}

		// if continuationToken has a value, then there is at least one more page of projects to get
		if responseValue.ContinuationToken != "" {

			continuationToken, err := strconv.Atoi(responseValue.ContinuationToken)
			if err != nil {
				log.Fatal(err)
			}

			// Get next page of team projects
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &continuationToken,
			}
			responseValue, err = coreClient.GetProjects(ctx, projectArgs)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			responseValue = nil
		}
	}
}

func GetCoreClientAndCtx(config Configuration) (core.Client, context.Context, error) {
	organizationUrl := "https://dev.azure.com/" + config.OrganizationName // todo: replace value with your organization url

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, config.PersonalAccessToken)

	ctx := context.Background()

	// Create a client to interact with the Core area
	coreClient, err := core.NewClient(ctx, connection)

	if err != nil {
		return coreClient, ctx, err
	}

	return coreClient, ctx, nil
}

func GetWorkClientAndCtx(config Configuration) (work.Client, context.Context, error) {
	organizationUrl := "https://dev.azure.com/" + config.OrganizationName // todo: replace value with your organization url

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, config.PersonalAccessToken)

	ctx := context.Background()

	// Create a client to interact with the Work area
	Client, err := work.NewClient(ctx, connection)

	if err != nil {
		return Client, ctx, err
	}

	return Client, ctx, nil
}


func GetWorkItemTrackingClientAndCtx(config Configuration) (workitemtracking.Client, context.Context, error) {
	organizationUrl := "https://dev.azure.com/" + config.OrganizationName // todo: replace value with your organization url

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, config.PersonalAccessToken)

	ctx := context.Background()

	// Create a client to interact with the Work area
	Client, err := workitemtracking.NewClient(ctx, connection)

	if err != nil {
		return Client, ctx, err
	}

	return Client, ctx, nil
}

func GetTeamUuid(config Configuration) (id uuid.UUID, err error) {
	WorkClient, ctx, err := GetWorkClientAndCtx(config)
	fmt.Printf("%v, %v, %v\n", WorkClient, ctx, err)
	CoreClient, ctx, err := GetCoreClientAndCtx(config)
	if err != nil{
		return id, err
	}
    WebApiTeams, err := CoreClient.GetAllTeams(ctx, core.GetAllTeamsArgs{})
	if err != nil{
		return id, err
	}

	for _, WebApiTeam := range *WebApiTeams{
		if config.TeamName == *WebApiTeam.Name {
			return *WebApiTeam.Id, nil
		}
	}
	
	return id, err
}

func GetIterationUuid(config Configuration) (id uuid.UUID, err error) {
	WorkClient, ctx, err := GetWorkClientAndCtx(config)
	fmt.Printf("%v, %v, %v\n", WorkClient, ctx, err)
	//CoreClient, ctx, err := GetCoreClientAndCtx(config)
	if err != nil{
		return id, err
	}

	/*teamID, err := GetTeamUuid(config)
	if err != nil{
		return id, err
	}*/

	//stringTeamID := teamID.String()
    TeamSettingsIterations, err := WorkClient.GetTeamIterations(ctx, work.GetTeamIterationsArgs{Project: &(config.ProjectName)})
	
	if TeamSettingsIterations == nil{
		log.Fatalf("failed to find iteration id : %v\n", TeamSettingsIterations)
		return id, err
	}
	for _, TeamSettingsIteration := range *TeamSettingsIterations{
	 fmt.Printf("Iteration: %v, %v\n", TeamSettingsIteration.Id, *TeamSettingsIteration.Name)
	 if *TeamSettingsIteration.Name == config.SprintName{
		return *TeamSettingsIteration.Id, nil
	 }
    }
	return id, err
}

func DownloadSprintStatus(config Configuration) error {
	WorkItemTrackingClient, ctx, err := GetWorkItemTrackingClientAndCtx(config)
	fmt.Printf("%v, %v, %v\n", WorkItemTrackingClient, ctx, err)
	
	/*IterationId, err:= GetIterationUuid(config)
	if err != nil{
		return err
	}
	
	teamID, err := GetTeamUuid(config)
	if err != nil{
		return err
	}

	stringTeamID := teamID.String()*/
	query := "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project"
	queryResult, err := WorkItemTrackingClient.QueryByWiql(context.Background(), workitemtracking.QueryByWiqlArgs{
		Wiql: &workitemtracking.Wiql{Query: &query},
		Project: &config.ProjectName,
	})
	
	ids := make([]int, 200)
	for i, wit := range (*queryResult.WorkItems)[:200]{
		ids[i] = *wit.Id
	}
	//fmt.Printf("queryResult.WorkItems: %v\n", *(.Id)

	if err != nil {
		log.Fatalf("Failed to fetch work item IDs: %v", err)
	}

    
	args := workitemtracking.GetWorkItemsArgs{Project: &config.ProjectName, Ids: &ids}
	IterationWorkItems, err := WorkItemTrackingClient.GetWorkItems(ctx, args)
	if err != nil{
		return err
	}
     
	fmt.Printf("IterationWorkItems: %v\n", IterationWorkItems)
	err = CacheToFile(IterationWorkItems)
	if err != nil{
		return err
	}
	return nil
}

func CacheToFile(IterationWorkItems *[]workitemtracking.WorkItem) (err error){
	jsonData, err := json.MarshalIndent(IterationWorkItems, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile("/tmp/wips.json", jsonData, 0644)
	if err != nil {
		return err
	}
	return nil

}

func SubmitSprintStatus(config Configuration, wobjects []Wobject) error {
	return fmt.Errorf("todo: %s, %s", config, wobjects)
}

func LoadConfig(configFilePath string) (config Configuration, err error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
