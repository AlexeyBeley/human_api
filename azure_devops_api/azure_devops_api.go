package azure_devops_api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
	ProjectName         string `json:"ProjectName"`
	SprintName          string `json:"SprintName"`
}

type Wobject struct {
	Id string
}

type WorkItem struct {
	ID        int                    `json:"id"`
	Rev       int                    `json:"rev"`
	Fields    map[string]interface{} `json:"fields"`
	Relations []struct {
		Rel        string                 `json:"rel"`
		URL        string                 `json:"url"`
		Attributes map[string]interface{} `json:"attributes"`
	}
}

func HoreyClient(config Configuration) error {
	// Azure DevOps organization and project details
	organization := config.OrganizationName
	project := config.OrganizationName
	workItemID := 11111              // Replace with the actual work item ID
	pat := config.PersonalAccessToken // Replace with your actual PAT

	// Construct the API URL
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/wit/workitems/%d?api-version=7.0&$expand=relations", organization, project, workItemID)

	// Create an HTTP client with a timeout (optional but recommended)
	client := http.Client{Timeout: 10 * time.Second}

	// Create a new request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set the Authorization header with your personal access token (PAT)

	base64Pat := basicAuth(pat)
	req.Header.Set("Authorization", "Basic "+base64Pat)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode the JSON response
	var workItem WorkItem
	err = json.NewDecoder(resp.Body).Decode(&workItem)
	if err != nil {
		log.Fatal(err)
	}

	// Access and print the relations
	fmt.Println("Relations:")
	for _, relation := range workItem.Relations {
		fmt.Printf("- %s: %s\n", relation.Rel, relation.URL)
		// You can access relation attributes using relation.Attributes
	}
	return nil
}

type witWorkItemRelation struct {
	Rel        *string                `json:"rel"`
	Url        *string                `json:"url"`
	Attributes map[string]interface{} `json:"attributes"`
}

type witWorkItemQueryResult struct {
	WorkItems         *[]witWorkItemReference `json:"workItems"`
	WorkItemRelations *[]witWorkItemRelation  `json:"workItemRelations"`
}

type witWorkItemReference struct {
	Id *int `json:"id"`
}

func GetAllWits(config Configuration) error {
	ctx := context.Background()

	// Fetch work item IDs in batches using WIQL
	ids, err := getWorkItemIDs(config, ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fetched %d", len(ids))
	return nil
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

// Helper function to create basic authentication header
func basicAuth(pat string) string {
	return base64.StdEncoding.EncodeToString([]byte(":" + pat))
}

func GetWorkClientAndCtx(config Configuration) (work.Client, context.Context, error) {
	organizationUrl := "https://dev.azure.com/" + config.OrganizationName

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

func ValidateConfig(config Configuration) error {
	if config.OrganizationName == "" {
		return fmt.Errorf("parameter OrganizationName was not set in config")
	}
	return nil
}

func GetWorkItemTrackingClientAndCtx(config Configuration) (workitemtracking.Client, context.Context, error) {
	err := ValidateConfig(config)
	if err != nil {
		return nil, nil, err
	}
	organizationUrl := "https://dev.azure.com/" + config.OrganizationName

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, config.PersonalAccessToken)

	ctx := context.Background()
	if ctx == nil {
		log.Fatal("Can not allocate context.Background")
	}

	// Create a client to interact with the Work area
	Client, err := workitemtracking.NewClient(ctx, connection)

	if err != nil {
		log.Printf("was not able to create new workitemtracking client")
		return Client, ctx, err
	}

	return Client, ctx, nil
}

func GetTeamUuid(config Configuration) (id uuid.UUID, err error) {
	WorkClient, ctx, err := GetWorkClientAndCtx(config)
	fmt.Printf("%v, %v, %v\n", WorkClient, ctx, err)
	CoreClient, ctx, err := GetCoreClientAndCtx(config)
	if err != nil {
		return id, err
	}
	WebApiTeams, err := CoreClient.GetAllTeams(ctx, core.GetAllTeamsArgs{})
	if err != nil {
		return id, err
	}

	for _, WebApiTeam := range *WebApiTeams {
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
	if err != nil {
		return id, err
	}

	/*teamID, err := GetTeamUuid(config)
	if err != nil{
		return id, err
	}*/

	//stringTeamID := teamID.String()
	TeamSettingsIterations, err := WorkClient.GetTeamIterations(ctx, work.GetTeamIterationsArgs{Project: &(config.ProjectName)})

	if TeamSettingsIterations == nil {
		log.Fatalf("failed to find iteration id : %v\n", TeamSettingsIterations)
		return id, err
	}
	for _, TeamSettingsIteration := range *TeamSettingsIterations {
		fmt.Printf("Iteration: %v, %v\n", TeamSettingsIteration.Id, *TeamSettingsIteration.Name)
		if *TeamSettingsIteration.Name == config.SprintName {
			return *TeamSettingsIteration.Id, nil
		}
	}
	return id, err
}



func CallGetWorkItems(config Configuration, ctx context.Context, WorkItemTrackingClient workitemtracking.Client, WitIds []int, ch chan *[]workitemtracking.WorkItem) (err error) {
	Fields := []string{"System.State", "System.Id", "System.CreatedBy", "System.CreatedDate"}
	args := workitemtracking.GetWorkItemsArgs{Project: &config.ProjectName, Ids: &WitIds, Fields: &Fields}
	IterationWorkItems, err := WorkItemTrackingClient.GetWorkItems(ctx, args)
	if err != nil {
		return err
	}

	ch <- IterationWorkItems
	close(ch)
	return nil
}

func GetAllFields() error{
	//todo: replace with real implementation
	connection := azuredevops.NewPatConnection("organizationUrl", "config.PersonalAccessToken")

	ctx := context.Background()

	// Create a client to interact with the Core area
	WorkItemTrackingClient, err := workitemtracking.NewClient(ctx, connection)
	variable := "&config.ProjectName"
	argsNew := workitemtracking.GetWorkItemFieldsArgs{Project: &variable}
	WorkItemField2, err := WorkItemTrackingClient.GetWorkItemFields(ctx, argsNew)
	if err != nil {
		return err
	}
	log.Printf("WorkItemField2: %v", (*WorkItemField2)[0])
	return nil
}
func CacheToFile(IterationWorkItems *[]workitemtracking.WorkItem, dstFilePath string) (err error) {
	jsonData, err := json.MarshalIndent(IterationWorkItems, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(dstFilePath, jsonData, 0644)
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

func getWorkItemIDs(config Configuration, ctx context.Context) ([]int, error) {

	client := http.Client{Timeout: 10 * time.Second}
	requestUrl := "https://dev.azure.com/" + config.OrganizationName + "/" +config.ProjectName + "/_apis/wit/wiql?api-version=7.0"
	wiqlData := fmt.Sprintf(`{"query": "SELECT [System.Id] FROM WorkItems Where [System.TeamProject] = '%s'"}`, config.ProjectName)
	AuthHeaderValue := "Basic "+basicAuth(config.PersonalAccessToken)
	
	jsonBody := []byte(wiqlData)
 	bodyReader := bytes.NewReader(jsonBody)
		
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestUrl, bodyReader)
	if err != nil {
		return []int{}, err
	}

	// Set the Authorization header
	req.Header.Set("Authorization", AuthHeaderValue)
	req.Header.Set("Content-Type", "application/json")
		
		// Send the request

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

		// Decode the JSON response
	var queryResult witWorkItemQueryResult
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	if err != nil {
		return nil, err
	}

	// Extract work item IDs
	lenIds := len(*queryResult.WorkItems)
	//allIDs := [lenIds]int[]{}
	var allIDs [20000]int
	if queryResult.WorkItems != nil {
		for i, workItem := range *queryResult.WorkItems {
			allIDs[i] = *workItem.Id
		}
	} else {
		log.Fatal("Can not fetch work item ids")
	}

	// Check if there are more results
	if queryResult.WorkItemRelations != nil && len(*queryResult.WorkItemRelations) != 0 {
		log.Fatal("Unexpected status: Length of the WorkItemRelations is not 0")
	}
		
	return allIDs[0:lenIds], nil
}
func getClient()http.Client{
	return http.Client{Timeout: 10 * time.Second}
}

func getRequest(config Configuration, ctx context.Context,  RequestPath string) (*http.Request, error){
	
	requestUrl := "https://dev.azure.com/" + config.OrganizationName + "/" +config.ProjectName + "/_apis/"+ RequestPath
	AuthHeaderValue := "Basic "+basicAuth(config.PersonalAccessToken)
	
		
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestUrl, nil)
	if err != nil {
		return req, err
	}

	// Set the Authorization header
	req.Header.Set("Authorization", AuthHeaderValue)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
func DownloadAllWits(config Configuration, dstFilePath string) error {
	ctx := context.Background()

	// Fetch work item IDs in batches using WIQL
	WitIds, err := getWorkItemIDs(config, ctx)
	if err != nil {
		log.Fatal(err)
	}

    //todo: remove 
	WitIds = WitIds[:400]
	//todo: end remove

	WitCount := len(WitIds)
	channelsCount := WitCount / 200

	channels := make([]chan *[]workitemtracking.WorkItem, channelsCount)
	for i := range channels {
		channels[i] = make(chan *[]workitemtracking.WorkItem)
	}

	for i := range channels {
		WitIdsSlice := WitIds[i*200 : i*200+200]
		go func() {
			GetWorkItemsBySlice(config, ctx, WitIdsSlice, channels[i])
		}()
	}

	//fmt.Printf("queryResult.WorkItems: %v\n", *(.Id)
	//*[]workitemtracking.WorkItem

	AllWits := []workitemtracking.WorkItem{}
	for _, ch := range channels {
		IterationWorkItems := <-ch
		AllWits = append(AllWits, *IterationWorkItems...)

	}
	fmt.Printf("IterationWorkItems: %d\n", len(AllWits))
	err = CacheToFile(&AllWits, dstFilePath)
	if err != nil {
		return err
	}
	return nil
}

func GetWorkItemsBySlice(config Configuration, ctx context.Context, WitIds []int, ch chan *[]workitemtracking.WorkItem)error{

	req, err := getRequest(config, ctx, fmt.Sprintf("wit/workitems/%d?$expand=all&api-version=7.0", WitIds[0]))
	if err != nil{
		return err
	}
	client := getClient()
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode the JSON response
	//var queryResult witWorkItemQueryResult
	var wits []workitemtracking.WorkItem
	err = json.NewDecoder(resp.Body).Decode(&wits)
	if err != nil {
		return err
	}
	return nil
}