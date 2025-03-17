package human_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AlexeyBeley/human_api/azure_devops_api"
)

type Configuration struct {
	SprintName                       string `json:"SprintName"`
	ReportsDirPath                   string `json:"ReportsDirPath"`
	WorkerId                         string `json:"WorkerId"`
	AzureDevopsConfigurationFilePath string `json:"AzureDevopsConfigurationFilePath"`
}

type Wobject struct {
	Id           string   `json:"Id"`
	Title        string   `json:"Title"`
	Description  string   `json:"Description"`
	LeftTime     int      `json:"LeftTime"`
	InvestedTime int      `json:"InvestedTime"`
	WorkerID     string   `json:"WorkerID"`
	ChildrenIDs  []string `json:"ChildrenIDs"`
	ParentID     string   `json:"ParentID"`
	Priority     int      `json:"Priority"`
	Status       string   `json:"Status"`
	Sprint       string   `json:"Sprint"`
	Type         string   `json:"Type"`
}

const preReportFileName = "pre_report.json"
const inputFileName = "input.hapi"
const baseFileName = "base.hapi"
const postReportFileName = "post_report.json"

func check(e error) {
	if e != nil {
		strErr := fmt.Sprintf("%v", e)
		data := []byte(strErr)
		err := os.WriteFile("/tmp/hapi.log", data, 0644) // 0644 are file permissions
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
		panic(e)
	}
}

func DailyRoutine(configFilePath string) error {
	/*
		if _, err:= os.Stat(reportFilePath) ; err == nil {
			fmt.Println("File exists")
		} else if os.IsNotExist(err) {
			fmt.Println("File does not exist")
		} else {
			fmt.Println("Error checking file existence:", err)
		}

	*/
	fmt.Println("starting daily routine")
	config, err := loadConfiguration(configFilePath)
	if err != nil {
		log.Printf("Failed with error: %v\n", err)
		return err
	}
	fmt.Println("Loaded config")

	now := time.Now()
	dateDirName := now.Format("2006_01_02")

	dateDirPath := filepath.Join(config.ReportsDirPath, config.SprintName, dateDirName)
	fmt.Println("Generated new directory path: " + dateDirPath)

	curDir, err := os.Getwd()
	check(err)
	fmt.Printf("Current workind dir: %v\n", curDir)

	os.Chdir(filepath.Join(config.ReportsDirPath, config.SprintName))

	//err = os.MkdirAll(filepath.Dir(dateDirPath), 0755)
	err = os.MkdirAll(dateDirName, 0755)
	if err != nil {
		fmt.Printf("was not able to create '%v'\n", dateDirPath)
		return err
	}
	os.Chdir(curDir)

	fmt.Println("Created new directory path: " + dateDirPath)

	preReportFilePath := filepath.Join(dateDirPath, preReportFileName)
	inputFilePath := filepath.Join(dateDirPath, inputFileName)
	baseFilePath := filepath.Join(dateDirPath, baseFileName)
	postReportFilePath := filepath.Join(dateDirPath, postReportFileName)

	if _, err := os.Stat(postReportFilePath); err == nil {
		return fmt.Errorf("post report file exists. The routine finished: %v", dateDirPath)
	}

	azure_devops_config, err := azure_devops_api.LoadConfig(config.AzureDevopsConfigurationFilePath)
	if err != nil {
		return err
	}
	log.Printf("inputFilePath: %v\n", inputFilePath)
	if !checkFileExists(inputFilePath) {
		return DailyRoutineExtract(config, azure_devops_config, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath)
	}
	if !checkFileExists(preReportFilePath) ||
		!checkFileExists(inputFilePath) ||
		!checkFileExists(baseFilePath) ||
		checkFileExists(postReportFilePath) {
		return fmt.Errorf("undefined status: %s", postReportFilePath)
	}
	return DailyRoutineSubmit(azure_devops_config, inputFilePath, baseFilePath, postReportFilePath)

}

func DailyRoutineExtract(config Configuration, azureDevopsConfig azure_devops_api.Configuration, preReportFilePath, inputFilePath, baseFilePath, postReportFilePath string) (err error) {
	if !checkFileExists(preReportFilePath) {
		if checkFileExists(inputFilePath) {
			return fmt.Errorf("pre report file does not exist. Input file exists '%v'", inputFilePath)
		}
		if checkFileExists(baseFilePath) {
			return fmt.Errorf("pre report file does not exist. Base file exists '%v'", baseFilePath)
		}
		DownloadAllWits(azureDevopsConfig, preReportFilePath)
	}

	if !checkFileExists(inputFilePath) {

		GenerateDailyReport(config, preReportFilePath, baseFilePath)
		//_, err = ConvertDailyJsonToHR(dailyJSONFilePath, baseFilePath)
		//check(err)

		err = copyFile(baseFilePath, inputFilePath)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return err
		}
		return nil
	} else if checkFileExists(baseFilePath) {
		return fmt.Errorf("input file does not exist. Base file exists '%v'", baseFilePath)
	}

	if _, err := os.Stat(preReportFilePath); err == nil {
		fmt.Println("File exists")
	} else if os.IsNotExist(err) {
		fmt.Println("File does not exist")

		//ConvertToHapi(filepath.Dir(reportFilePath))
	} else {
		fmt.Println("Error checking file existence:", err)
	}
	return nil
}

func GenerateDailyReport(config Configuration, statusFilePath string, dstFilePath string) {
	wobjects, err := ConvertAzureDevopsStatusToWobjects(statusFilePath)
	check(err)
	GenerateDailyReportFromWobjects(config, wobjects, dstFilePath)
	//WorkerDailyReport{}
}

func GenerateDailyReportFromWobjects(config Configuration, wobjects []Wobject, dstFilePath string) (reportFilePath string) {
	log.Printf("filtering relevant wobkjects: %v\n", len(wobjects))
	wobjectsRelevant := FilterRelevantDailyReportWobjects(config, wobjects)
	new := []WorkerWobjReport{}
	active := []WorkerWobjReport{}
	blocked := []WorkerWobjReport{}
	closed := []WorkerWobjReport{}

	/*
	       Parent       []string `json:"parent"`
	   	Child        []string `json:"child"`
	   	Comment      string   `json:"comment"`
	   	InvestedTime int      `json:"invested_time"`
	   	LeftTime     int      `json:"left_time"`

	   	Parent (type, id, title)
	   	Child (type, id, title)
	*/

	reports := []WorkerDailyReport{}
	var workerID string
	for wobjid, wobject := range wobjectsRelevant {
		if wobjid == "-1" {
			continue
		}
		var parentPointer *Wobject
		var childPointer *Wobject

		if len(wobject.ChildrenIDs) != 0 {
			log.Printf("skipping Wobject with children: %v\n", wobjid)
			continue
		}

		if wobject.ParentID != "" {
			parentPointer = wobjectsRelevant[wobject.ParentID]
			childPointer = wobject
		} else {
			parentPointer, childPointer = GenerateParentAndChildFromParentlessWobject(wobject, wobjectsRelevant)
		}

		workerID = wobject.WorkerID

		report := WorkerWobjReport{Parent: []string{parentPointer.Type, parentPointer.Id, parentPointer.Title},
			Child: []string{childPointer.Type, childPointer.Id, childPointer.Title}}
		switch wobject.Status {
		case "New":
			new = append(new, report)
		case "Closed":
			closed = append(closed, report)
		case "Active":
			active = append(active, report)
		case "Blocked":
			blocked = append(blocked, report)
		default:
			check(fmt.Errorf("invalid wobject.Status: %v", wobject.Status))
		}
	}

	if len(new) == 0 {
		check(fmt.Errorf("new wobjects are empty: %v", new))
	}

	workerDailyReport := WorkerDailyReport{WorkerID: workerID,
		New:     new,
		Active:  active,
		Blocked: blocked,
		Closed:  closed,
	}
	reports = append(reports, workerDailyReport)
	WriteDailyToHRFile(reports, dstFilePath)
	return reportFilePath
}

func GenerateParentAndChildFromParentlessWobject(wobject *Wobject, wobjectsRelevant map[string]*Wobject) (parent, child *Wobject) {
	// For Task/Bug creates default parent "-1"
	// For other types creats default child "-"

	if wobject.Type == "Task" || wobject.Type == "Bug" {
		wobject.ParentID = "-1"
		wobjectsRelevant["-1"].ChildrenIDs = append(wobjectsRelevant["-1"].ChildrenIDs, wobject.Id)
		parent = wobjectsRelevant["-1"]
		child = wobject
	} else {
		wobject.ChildrenIDs = []string{"-1"}
		child = wobjectsRelevant["-1"]
		parent = wobject
	}

	return parent, child
}

func FilterRelevantDailyReportWobjects(config Configuration, wobjects []Wobject) map[string]*Wobject {

	log.Printf("filtering relevant wobjects: %v\n", len(wobjects))
	wobjectsRelevantById := make(map[string]*Wobject)
	wobjectsById := make(map[string]Wobject)

	for _, wobject := range wobjects {
		wobjectsById[wobject.Id] = wobject
	}

	for _, wobject := range wobjects {
		if wobject.WorkerID != config.WorkerId {
			continue
		}
		if wobject.Sprint != config.SprintName {
			continue
		}
		if _, exists := wobjectsRelevantById[wobject.Id]; exists {
			continue
		}
		wobjectsRelevantById[wobject.Id] = &wobject
		parent := wobjectsById[wobject.ParentID]
		parent.ChildrenIDs = append(parent.ChildrenIDs, wobject.Id)
		if _, exists := wobjectsRelevantById[wobject.ParentID]; !exists {
			wobjectsRelevantById[wobject.ParentID] = &parent
		}
	}

	wobjectsRelevantById["-1"] = &Wobject{Id: "-1",
		Description: "-1",
		Title:       "-1",
	}

	return wobjectsRelevantById
}

func ConvertAzureDevopsStatusToWobjects(filePath string) (wobjects []Wobject, err error) {
	wits, err := azure_devops_api.ReadWitsFromFile(filePath)

	check(err)
	//log.Printf("todo: %v\n", wits)
	for _, wit := range wits {
		wobject, err := ConvertWitToWobject(wit)
		check(err)
		wobjects = append(wobjects, wobject)
	}
	return wobjects, nil
}

func ConvertWitToWobject(wit azure_devops_api.WorkItem) (wobject Wobject, err error) {
	wobject.ParentID = extractFloat64String(wit, "System.Parent")
	wobject.Id = strconv.Itoa(wit.ID)
	wobject.Title = wit.Fields["System.Title"].(string)
	wobject.Priority = extractFloat64Int(wit, "Microsoft.VSTS.Common.Priority")

	wobject.WorkerID = extractWorkerID(wit)

	wobject.Status = extractStatus(wit)
	SprintParts := strings.Split(wit.Fields["System.IterationPath"].(string), "\\")
	wobject.Sprint = SprintParts[len(SprintParts)-1]
	wobject.Type = strings.Replace(wit.Fields["System.WorkItemType"].(string), " ", "", -1)
	return wobject, nil
}

func extractStatus(workItem azure_devops_api.WorkItem) string {
	SystemState := workItem.Fields["System.State"].(string)
	switch SystemState {
	case "New":
		return "New"
	case "Closed":
		return "Closed"
	case "Resolved":
		return "Closed"
	case "Removed":
		return "Closed"
	case "Active":
		return "Active"
	case "Blocked":
		return "Blocked"
	default:
		log.Printf("invalid State: %v, using default\n", SystemState)
		return "Blocked"
	}
}

func extractWorkerID(workItem azure_devops_api.WorkItem) string {
	var data string
	if workItem.Fields["System.AssignedTo"] != nil {
		data = workItem.Fields["System.AssignedTo"].(map[string]interface{})["uniqueName"].(string)
	} else {
		data = workItem.Fields["System.CreatedBy"].(map[string]interface{})["uniqueName"].(string)
	}

	return strings.Split(data, "@")[0]
}

func extractFloat64Int(workItem azure_devops_api.WorkItem, FieldKey string) int {
	var retVal int
	if workItem.Fields[FieldKey] == nil {
		return retVal
	}

	value, ok := workItem.Fields[FieldKey]
	if !ok {
		check(fmt.Errorf("extractFloat64Int: Was not able to Extract %v, %v, %v", FieldKey, value, workItem))
	}
	retVal, err := strconv.Atoi(strconv.FormatFloat(value.(float64), 'f', 0, 64))
	check(err)
	return retVal
}

func extractFloat64String(workItem azure_devops_api.WorkItem, FieldKey string) string {
	var retVal string
	if workItem.Fields[FieldKey] == nil {
		return retVal
	}

	value, ok := workItem.Fields[FieldKey]
	if !ok {
		check(fmt.Errorf("extractFloat64String: Was not able to Extract %v, %v, %v", FieldKey, value, workItem))
	}
	retValtmp, err := strconv.Atoi(strconv.FormatFloat(value.(float64), 'f', 0, 64))
	check(err)
	retVal = strconv.Itoa(retValtmp)
	return retVal
}

func DailyRoutineSubmit(config azure_devops_api.Configuration, inputFilePath, baseFilePath, postReportFilePath string) (err error) {
	inputWobjects := GetWobjectsFromReportFile(config, inputFilePath)
	baseWobjects := GetWobjectsFromReportFile(config, baseFilePath)

	err = CleanWobjectsUserInput(inputWobjects)
	if err != nil {
		return err
	}

	err = ValidateWobjectsUserInput(baseWobjects, inputWobjects)
	if err != nil {
		return err
	}

	wobjects := FilterChangedWobjects(baseWobjects, inputWobjects)

	wits := GenerateWorkItemsFromWobjects(wobjects)
	err = azure_devops_api.SubmitSprintStatus(config, wits)
	return err
}

func GetWobjectsFromReportFile(config azure_devops_api.Configuration, filePath string) (map[string]*Wobject) {
	inputJsonFilePath := strings.Replace(filepath.Base(filePath), ".hapi", "_hapi.json", 1)

	reports, err := ConvertHRToDailyJson(filePath, inputJsonFilePath)
	check(err)

	return GenerateWobjectsFromDailyReports(config, reports)

}

func CleanWobjectsUserInput(inputWobjects map[string]*Wobject) error {
	for _, wobject := range inputWobjects {
		cutset := " "
		wobject.Title = strings.TrimRight(strings.TrimLeft(wobject.Title, cutset), cutset)
		wobject.WorkerID = strings.TrimRight(strings.TrimLeft(wobject.WorkerID, cutset), cutset)
		wobject.Description = strings.TrimRight(strings.TrimLeft(wobject.Description, cutset), cutset)
	}

	return nil
}

func ValidateWobjectsUserInput(baseById map[string]*Wobject, inputWobjects map[string]*Wobject) error {
	errors := []string{}
	for _, wobject := range inputWobjects {

		cutset := "\t\r\n"
		cutset_readable := strings.ReplaceAll(cutset, "\t", "\\t, ")
		cutset_readable = strings.ReplaceAll(cutset_readable, "\n", "\\n, ")
		cutset_readable = strings.ReplaceAll(cutset_readable, "\r", "\\r, ")

		if strings.ContainsAny(wobject.Title, cutset) {
			errors = append(errors, fmt.Sprintf("wobject title '%s' contains one of invalid characters: [%s]", wobject.Title, cutset_readable))
		}

		cutset = "\t \r\n"
		cutset_readable = strings.ReplaceAll(cutset, " ", "\\s, ")
		cutset_readable = strings.ReplaceAll(cutset_readable, "\t", "\\t, ")
		cutset_readable = strings.ReplaceAll(cutset_readable, "\n", "\\n, ")
		cutset_readable = strings.ReplaceAll(cutset_readable, "\r", "\\r, ")
		if strings.ContainsAny(wobject.WorkerID, cutset) {
			errors = append(errors, fmt.Sprintf("wobject WorkerID '%s' contains one of invalid characters: [%s]", wobject.WorkerID, cutset_readable))
		}

		if wobject.Id != "" {
			_, ok := baseById[wobject.Id]
			if !ok {
				errors = append(errors, fmt.Sprintf("wobject Id '%s' from input does not exist in base file", wobject.Id))
			} else {
				errors = append(errors, ValidateWobjectUserInput(wobject)...)
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("input Validation errors:\n %v", strings.Join(errors, "\n"))
	}
	return nil
}

func ValidateWobjectUserInput(wobject *Wobject) (errors []string) {
	//new task/bug wobject
	if wobject.Id == "-1" {
		return errors
	}

	if len(wobject.ChildrenIDs) == 0 {
		if wobject.Type != "Task" && wobject.Type != "Bug" {
			errors = append(errors, fmt.Sprintf("[%s][%s] - unsupported Wobject Type %s. Use one of ['Task', 'Bug']", wobject.Id, wobject.Title, wobject.Type))
		}
	}

	return errors
}

func FilterChangedWobjects(baseById map[string]*Wobject, inputWobjects map[string]*Wobject) (wobjectsRet []*Wobject) {
	for _, inputWobject := range inputWobjects {
		if inputWobject.Id == "-1" {
			check(fmt.Errorf("FilterChangedWobjects does not expect to get unparented object: %v", inputWobject))
		}
		//New wobject
		if inputWobject.Id == "" {
			wobjectsRet = append(wobjectsRet, inputWobject)
			continue
		}

		baseWobject, ok := baseById[inputWobject.Id]
		if !ok {
			check(fmt.Errorf("input Wobject ID '%v' does not exist in base.haphi ", inputWobject.Id))
			continue
		}

		if inputWobject.Description == baseWobject.Description &&
			inputWobject.InvestedTime == baseWobject.InvestedTime &&
			inputWobject.LeftTime == baseWobject.LeftTime &&
			inputWobject.Status == baseWobject.Status {
			continue
		}
		wobjectsRet = append(wobjectsRet, inputWobject)
	}
	return wobjectsRet
}

func GenerateWobjectsFromDailyReports(cofig azure_devops_api.Configuration, reports []WorkerDailyReport) (map[string]*Wobject) {
	wobjectById := make(map[string]*Wobject)
	for _, report := range reports {
		for _, wobjectReport := range report.New {
			GenerateWobjectsFromWobjectReport(cofig, wobjectById, report.WorkerID, "New", wobjectReport)
		}

		for _, wobjectReport := range report.Active {
			GenerateWobjectsFromWobjectReport(cofig, wobjectById, report.WorkerID, "Active", wobjectReport)
		}

		for _, wobjectReport := range report.Blocked {
			GenerateWobjectsFromWobjectReport(cofig, wobjectById, report.WorkerID, "Blocked", wobjectReport)
		}
		for _, wobjectReport := range report.Closed {
			GenerateWobjectsFromWobjectReport(cofig, wobjectById, report.WorkerID, "Closed", wobjectReport)
		}
	}
	return wobjectById
}

func GenerateWobjectsFromWobjectReport(cofig azure_devops_api.Configuration, wobjectById map[string]*Wobject, WorkerID string, status string, wobjectReport WorkerWobjReport) {
	//{type, id, title}

	if wobjectReport.Parent[1] != "-1" {
		if _, ok := wobjectById[wobjectReport.Parent[1]]; !ok {
			wobj := Wobject{Id: wobjectReport.Parent[1],
				Title:        wobjectReport.Parent[2],
				WorkerID:     WorkerID,
				ChildrenIDs:  []string{},
				Priority:     -1,
				InvestedTime: -1,
				LeftTime:     wobjectReport.LeftTime,
				Status:       status,
				Sprint:       cofig.SprintName,
				Type:         wobjectReport.Parent[0],
			}
			wobjectById[wobj.Id] = &wobj
		}
	}

	if value, ok := wobjectById[wobjectReport.Child[1]]; ok {
		if value.Id == "-1"{
			return
		}
		check(fmt.Errorf("reported child wobject ID '%v' already appeared in a report with title %v", value.Id, value.Title))
	}

	wobj := Wobject{Id: wobjectReport.Child[1],
		Title:        wobjectReport.Child[2],
		WorkerID:     WorkerID,
		ChildrenIDs:  []string{},
		Priority:     -1,
		Status:       status,
		Sprint:       cofig.SprintName,
		InvestedTime: wobjectReport.InvestedTime,
		LeftTime:     wobjectReport.LeftTime,
		Description:  wobjectReport.Comment,
		Type:         wobjectReport.Child[0],
		ParentID:     wobjectReport.Parent[1],
	}

	wobjectById[wobj.Id] = &wobj
}

func GenerateWorkItemsFromWobjects(wobjects []*Wobject) (wits []azure_devops_api.WorkItem) {
	for _, wobject := range wobjects {
		witId := 0
		var err error

		if wobject.Id != "" {
			witId, err = strconv.Atoi(wobject.Id)
			check(err)
		}

		wit := azure_devops_api.WorkItem{
			ID: witId,
		}
		wits = append(wits, wit)
	}

	return wits
}

func loadConfiguration(filePath string) (config Configuration, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func DownloadAllWits(config azure_devops_api.Configuration, dstFilePath string) (err error) {
	log.Printf("downloadAllWits: %v, %v\n", config, dstFilePath)
	err = azure_devops_api.DownloadAllWits(config, dstFilePath)
	return err
}

// Return True if exists, False if not or fails on error.
func checkFileExists(path string) (exists bool) {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	log.Fatalf("Failed checking file exists: %v", err)
	return false
}

func copyFile(srcFilePath, dstFilePath string) error {
	// Open the source file for reading
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println("Error opening source file:", err)
		return err
	}
	defer srcFile.Close()

	// Create the destination file (with 0644 permissions)
	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		fmt.Println("Error creating destination file:", err)
		return err
	}
	defer dstFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return err
	}

	return nil
}

func logWithLineNumber(message string) {
	// Get the caller's file name and line number
	_, file, line, ok := runtime.Caller(1) // 1 skips the current function
	if !ok {
		file = "???"
		line = 0
	}

	// Format the log message with line number
	logMessage := fmt.Sprintf("%s:%d: %s", file, line, message)

	// Print the log message
	fmt.Println(logMessage)
}
