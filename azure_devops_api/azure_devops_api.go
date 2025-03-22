package azure_devops_api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	AreaPath            string `json:"AreaPath"`
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
	workItemID := 11111               // Replace with the actual work item ID
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
	fmt.Printf("fetched %d\n", len(ids))
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


func GetIteration(config Configuration) (iteration work.TeamSettingsIteration, err error) {
	WorkClient, ctx, err := GetWorkClientAndCtx(config)

	if err != nil {
		return iteration, err
	}

	TeamSettingsIterations, err := WorkClient.GetTeamIterations(ctx, work.GetTeamIterationsArgs{Project: &(config.ProjectName)})

	if TeamSettingsIterations == nil {
		return iteration, err
	}
	for _, TeamSettingsIteration := range *TeamSettingsIterations {
		if *TeamSettingsIteration.Name == config.SprintName {
			return TeamSettingsIteration, nil
		}
	}
	return iteration, fmt.Errorf("was not able to find Iteration by name: %s", config.SprintName)
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

func GetAllFields() error {
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

	systemAreaId := "todo: fetch before the request"
	panic("todo:")
	client := http.Client{Timeout: 10 * time.Second}
	requestUrl := "https://dev.azure.com/" + config.OrganizationName + "/" + config.ProjectName + "/_apis/wit/wiql?api-version=7.0"
	wiqlData := fmt.Sprintf(`{"query": "SELECT [System.Id] FROM WorkItems Where [System.TeamProject] = '%s' AND [System.AreaId] = %s"}`, config.ProjectName, systemAreaId)
	AuthHeaderValue := "Basic " + basicAuth(config.PersonalAccessToken)

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
func getClient() http.Client {
	return http.Client{Timeout: 10 * time.Second}
}

func createRequest(config Configuration, ctx context.Context, RequestPath string, httpMethod string, body io.Reader, contentType string) (*http.Request, error) {

	requestUrl := "https://dev.azure.com/" + config.OrganizationName + "/" + config.ProjectName + "/_apis/" + RequestPath
	AuthHeaderValue := "Basic " + basicAuth(config.PersonalAccessToken)

	req, err := http.NewRequestWithContext(ctx, httpMethod, requestUrl, body)
	if err != nil {
		return req, err
	}

	// Set the Authorization header
	req.Header.Set("Authorization", AuthHeaderValue)
	req.Header.Set("Content-Type", contentType)
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
	//WitIds = WitIds[:400]
	//todo: end remove
	BulckSize := 50
	WitCount := len(WitIds)
	channelsCount := WitCount / BulckSize
	if BulckSize*channelsCount < WitCount {
		channelsCount += 1
	}

	channels := make([]chan *[]workitemtracking.WorkItem, channelsCount)
	for chanIndex := range channels {
		channels[chanIndex] = make(chan *[]workitemtracking.WorkItem)
	}

	i := 0
	for i < WitCount {
		log.Printf("Entering loop with i: %d\n", i)
		endIndex := i + BulckSize
		log.Printf("loop i:%d, endIndex: %d\n", i, endIndex)
		if i+BulckSize >= WitCount {
			endIndex = WitCount - 1
		}

		log.Printf("loop i:%d, endIndex after cahnge: %d\n", i, endIndex)
		log.Printf("loop i: %d, endIndex:%d, i/BulckSize: %d\n", i, endIndex, i/BulckSize)
		if i/BulckSize == 8 {
			log.Printf("Problem loop i: %d, endIndex:%d, i/BulckSize: %d\n", i, endIndex, i/BulckSize)

		}
		WitIdsSlice := WitIds[i:endIndex]
		chanIndex := i / BulckSize
		go func() {
			GetWorkItemsBySlice(config, ctx, WitIdsSlice, channels[chanIndex])
		}()

		i += BulckSize
	}

	//fmt.Printf("queryResult.WorkItems: %v\n", *(.Id)
	//*[]workitemtracking.WorkItem

	AllWits := []workitemtracking.WorkItem{}
	for j, ch := range channels {
		fmt.Printf("fetched from chanel %d out of %d channels\n", j, len(channels))
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

func GetWorkItemsBySlice(config Configuration, ctx context.Context, WitIds []int, ch chan *[]workitemtracking.WorkItem) error {
	retWorkItems := []workitemtracking.WorkItem{}

	for i, WitId := range WitIds {
		fmt.Printf("fetched witid  : %d/%d\n", i, len(WitIds))
		req, err := createRequest(config, ctx, fmt.Sprintf("wit/workitems/%d?$expand=all&api-version=7.0", WitId), http.MethodGet, nil, "application/json")
		if err != nil {
			return err
		}
		client := getClient()

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("received error in HTTP clinet request: %v", err)
			return err
		}
		defer resp.Body.Close()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
		}

		// Decode the JSON response
		//var queryResult witWorkItemQueryResult
		var wit workitemtracking.WorkItem
		err = json.NewDecoder(resp.Body).Decode(&wit)
		if err != nil {
			return err
		}
		retWorkItems = append(retWorkItems, wit)
	}

	ch <- &retWorkItems
	return nil
}

func ReadWitsFromFile(filePath string) (wits []WorkItem, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &wits)
	if err != nil {
		return nil, err
	}
	return wits, nil
}

func SubmitSprintStatus(config Configuration, requestDicts []map[string]string) error {
	// Provision parents
	// todo: Clean new identical parents by title
	for _, requestDict := range requestDicts {
		if requestDict["ChildrenIDs"] == "" {
			continue
		}
		err := ProvisionWitFromDict(config, requestDict)
		if err != nil {
			return nil
		}
	}

	for _, requestDict := range requestDicts {
		if requestDict["ChildrenIDs"] != "" {
			continue
		}

		err := ProvisionWitFromDict(config, requestDict)
		if err != nil {
			return nil
		}
	}
	return nil
}

func ProvisionWitFromDict(config Configuration, requestDict map[string]string) error {
	// provision_work_item_from_dict

	if requestDict["Id"] == "-1" {
		return nil
	}

	if strings.HasPrefix(requestDict["Id"], "CreatePlease:") {
		return CreateWit(config, requestDict)
	}
	return UpdateWit(config, requestDict)
}

func CreateWit(config Configuration, requestDict map[string]string) error {
	req, err := GenerateCreateWitRequest(config, requestDict)
	if err != nil {
		log.Printf("received error in Generate Create Wit Request: %v", err)
		return err
	}
	client := getClient()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("received error in HTTP clinet request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode the JSON response
	//var queryResult witWorkItemQueryResult
	var wit workitemtracking.WorkItem
	err = json.NewDecoder(resp.Body).Decode(&wit)
	if err != nil {
		return err
	}

	return nil
}

func GenerateCreateWitRequest(config Configuration, requestDict map[string]string) (*http.Request, error) {
	/*
		dictRequest["Id"] = wobject.Id
		dictRequest["ParentID"] = wobject.ParentID
		dictRequest["Priority"] = strconv.Itoa(wobject.Priority)
		dictRequest["Title"] = wobject.Title
		dictRequest["Description"] = wobject.Description
		dictRequest["LeftTime"] = strconv.Itoa(wobject.LeftTime)
		dictRequest["InvestedTime"] = strconv.Itoa(wobject.InvestedTime)
		dictRequest["WorkerID"] = wobject.WorkerID
		dictRequest["ChildrenIDs"] = strings.Join(*wobject.ChildrenIDs, ",")
		dictRequest["Sprint"] = wobject.Sprint
		dictRequest["Status"] = wobject.Status
		dictRequest["Type"] = wobject.Type
	*/
	
	if config.AreaPath == ""{
		return nil, fmt.Errorf("error config.AreaPath is empty, %v", config)
	}

	if value, err := strconv.Atoi(requestDict["Priority"]); err != nil || value== -1 {
		return nil, fmt.Errorf("creating Wobject has malformed Prioriy: %v, %v", value, err)
	}

	ctx := context.Background()
	postList := []map[string]string{}

	var witUrlType string
	switch {
	case requestDict["Type"] == "UserStory":
		witUrlType = "$User%20Story"
	case requestDict["Type"] == "Task" || requestDict["Type"] == "Bug":
		witUrlType = "$" + requestDict["Type"]
	default:
		return nil, fmt.Errorf("unknown WIT Type: %s", requestDict["Type"])
	}

	postData, err := json.Marshal(postList)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.AreaPath",
		"value": config.AreaPath,
	})

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.Title",
		"value": requestDict["Title"],
	})

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.Description",
		"value": requestDict["Description"],
	})

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/Microsoft.VSTS.Common.Priority",
		"value": requestDict["Priority"],
	})

	iteration, err := GetIteration(config)
	if err != nil{
		return nil, err
	}

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.IterationPath",
		"value": *iteration.Path,
	})

	err = fillCreateWitRequestTimes(&postList, requestDict)
	if err != nil{
		return nil, err
	}
	
	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.AssignedTo",
		"value": requestDict["WorkerID"],
	})


	fmt.Printf("Creating new Azure Devops WorkITem  : %v\n", requestDict)

	req, err := createRequest(config, ctx, fmt.Sprintf("wit/workitems/%s?api-version=7.0", witUrlType), http.MethodPost, bytes.NewBuffer(postData), "application/json-patch+json")
	return req, err
}

func fillCreateWitRequestTimes(postList *[]map[string]string, requestDict map[string]string) (error) {
	if requestDict["LeftTime"] == "-1"{
		return nil
	}

	*postList = append(*postList, map[string]string{
		"op":    "add",
		"path":  "/fields/Microsoft.VSTS.Scheduling.RemainingWork",
		"value": requestDict["LeftTime"],
	})

	*postList = append(*postList, map[string]string{
		"op":    "add",
		"path":  "/fields/Microsoft.VSTS.Scheduling.CompletedWork",
		"value": requestDict["InvestedTime"],
	})

	intLeftTime, err := strconv.Atoi(requestDict["LeftTime"])
	if err != nil {
		return err
	}
	
	intInvestedTime, err := strconv.Atoi(requestDict["InvestedTime"])	
	if err != nil {
		return err
	}

	originalEstimate := strconv.Itoa(intLeftTime + intInvestedTime)

	*postList = append(*postList, map[string]string{
		"op":    "add",
		"path":  "/fields/Microsoft.VSTS.Scheduling.OriginalEstimate",
		"value": originalEstimate,
	})
	return nil
}

func UpdateWit(config Configuration, requestDict map[string]string) error {
	req, err := GenerateUpdateWitRequest(config, requestDict)
	if err != nil {
		log.Printf("received error in Generate Create Wit Request: %v", err)
		return err
	}
	client := getClient()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("received error in HTTP clinet request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status error: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode the JSON response
	//var queryResult witWorkItemQueryResult
	var wit workitemtracking.WorkItem
	err = json.NewDecoder(resp.Body).Decode(&wit)
	if err != nil {
		return err
	}

	return nil
}

func GenerateUpdateWitRequest(config Configuration, requestDict map[string]string) (*http.Request, error) {
	/*
		dictRequest["Id"] = wobject.Id
		dictRequest["ParentID"] = wobject.ParentID
		dictRequest["Priority"] = strconv.Itoa(wobject.Priority)
		dictRequest["Title"] = wobject.Title
		dictRequest["Description"] = wobject.Description
		dictRequest["LeftTime"] = strconv.Itoa(wobject.LeftTime)
		dictRequest["InvestedTime"] = strconv.Itoa(wobject.InvestedTime)
		dictRequest["WorkerID"] = wobject.WorkerID
		dictRequest["ChildrenIDs"] = strings.Join(*wobject.ChildrenIDs, ",")
		dictRequest["Sprint"] = wobject.Sprint
		dictRequest["Status"] = wobject.Status
		dictRequest["Type"] = wobject.Type
	*/
	
	if config.AreaPath == ""{
		return nil, fmt.Errorf("error config.AreaPath is empty, %v", config)
	}

	ctx := context.Background()
	postList := []map[string]string{}

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.AreaPath",
		"value": config.AreaPath,
	})

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.Title",
		"value": requestDict["Title"],
	})

	if requestDict["Priority"] != "-1"{
		postList = append(postList, map[string]string{
			"op":    "add",
			"path":  "/fields/Microsoft.VSTS.Common.Priority",
			"value": requestDict["Priority"],
		})
	}

	iteration, err := GetIteration(config)
	if err != nil{
		return nil, err
	}

	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.IterationPath",
		"value": *iteration.Path,
	})


	err = fillUpdateWitRequestTimes(&postList, requestDict)
	if err != nil{
		return nil, err
	}
	
	postList = append(postList, map[string]string{
		"op":    "add",
		"path":  "/fields/System.AssignedTo",
		"value": requestDict["WorkerID"],
	})

	postData, err := json.Marshal(postList)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}
	fmt.Printf("Updating Azure Devops WorkITem  : %v\n", requestDict)

	req, err := createRequest(config, ctx, fmt.Sprintf("wit/workitems/%s?api-version=7.0", requestDict["Id"]), http.MethodPost, bytes.NewBuffer(postData), "application/json-patch+json")
	
	return req, err
}

func fillUpdateWitRequestTimes(postList *[]map[string]string, requestDict map[string]string) (error) {
	
	if requestDict["LeftTime"] != "-1"{
		_, err := strconv.Atoi(requestDict["LeftTime"])
		if err != nil {
			return err
		}
		
		*postList = append(*postList, map[string]string{
			"op":    "add",
			"path":  "/fields/Microsoft.VSTS.Scheduling.RemainingWork",
			"value": requestDict["LeftTime"],
		})
	}

	if requestDict["InvestedTime"] != "-1"{
		_, err := strconv.Atoi(requestDict["InvestedTime"])	
		if err != nil {
			return err
		}
		
		*postList = append(*postList, map[string]string{
			"op":    "add",
			"path":  "/fields/Microsoft.VSTS.Scheduling.CompletedWork",
			"value": requestDict["InvestedTime"],
		})
	}

	return nil
}