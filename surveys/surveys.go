package surveys

import (
	"encoding/json"
	"errors"
	"io"
	"log"

	"fmt"
	"sort"
	"strings"

	"github.com/AreaHQ/jsonhal"
	"github.com/ONSdigital/eq-questionnaire-launcher/clients"
	"github.com/ONSdigital/eq-questionnaire-launcher/oidc"
	"github.com/ONSdigital/eq-questionnaire-launcher/settings"
	"golang.org/x/oauth2"
)

// LauncherSchema is a representation of a schema in the Launcher
type LauncherSchema struct {
	Name       string
	SurveyType string
	URL        string
}

type DatasetMetadata struct {
	SurveyID            string `json:"survey_id"`
	PeriodID            string `json:"period_id"`
	Title               string `json:"title"`
	SdsSchemaVersion    int    `json:"sds_schema_version"`
	SdsPublishedAt      string `json:"sds_published_at"`
	TotalReportingUnits int    `json:"total_reporting_units"`
	SchemaVersion       string `json:"schema_version"`
	SdsDatasetVersion   int    `json:"sds_dataset_version"`
	Filename            string `json:"filename"`
	DatasetID           string `json:"dataset_id"`
}

// RegisterResponse is the response from the eq-survey-register request
type RegisterResponse struct {
	jsonhal.Hal
}

// Schemas is a list of Schema
type Schemas []Schema

// Schema is an available schema
type Schema struct {
	jsonhal.Hal
	Name string `json:"name"`
}

// LauncherSchemaFromFilename creates a LauncherSchema record from a schema filename
func LauncherSchemaFromFilename(filename string, surveyType string) LauncherSchema {
	return LauncherSchema{
		Name:       filename,
		SurveyType: surveyType,
	}
}

// GetAvailableSchemas Gets the list of static schemas an joins them with any schemas from the eq-survey-register if defined
func GetAvailableSchemas() map[string][]LauncherSchema {
	runnerSchemas := getAvailableSchemasFromRunner()
	registerSchemas := getAvailableSchemasFromRegister()

	allSchemas := append(runnerSchemas, registerSchemas...)

	sort.Sort(ByFilename(allSchemas))

	schemasBySurveyType := map[string][]LauncherSchema{}
	for _, schema := range allSchemas {
		schemasBySurveyType[strings.Title(schema.SurveyType)] = append(schemasBySurveyType[strings.Title(schema.SurveyType)], schema)
	}

	return schemasBySurveyType
}

// ByFilename implements sort.Interface based on the Name field.
type ByFilename []LauncherSchema

func (a ByFilename) Len() int           { return len(a) }
func (a ByFilename) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByFilename) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getAvailableSchemasFromRegister() []LauncherSchema {

	schemaList := []LauncherSchema{}

	if settings.Get("SURVEY_REGISTER_URL") != "" {
		resp, err := clients.GetHTTPClient().Get(settings.Get("SURVEY_REGISTER_URL"))
		if err != nil {
			log.Fatal("Do: ", err)
			return []LauncherSchema{}
		}

		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return schemaList
		}

		var registerResponse RegisterResponse
		if err := json.Unmarshal(responseBody, &registerResponse); err != nil {
			log.Print(err)
			return schemaList
		}

		var schemas Schemas

		schemasJSON, _ := json.Marshal(registerResponse.Embedded["schemas"])

		if err := json.Unmarshal(schemasJSON, &schemas); err != nil {
			log.Println(err)
		}

		for _, schema := range schemas {
			url := schema.Links["self"]
			schemaList = append(schemaList, LauncherSchema{
				Name:       schema.Name,
				URL:        url.Href,
				SurveyType: "Other",
			})
		}
	}

	return schemaList
}

func getAvailableSchemasFromRunner() []LauncherSchema {

	schemaList := []LauncherSchema{}

	hostURL := settings.Get("SURVEY_RUNNER_SCHEMA_URL")

	log.Printf("Survey Runner Schema URL: %s", hostURL)

	url := fmt.Sprintf("%s/schemas", hostURL)

	resp, err := clients.GetHTTPClient().Get(url)

	if err != nil {
		return []LauncherSchema{}
	}

	if resp.StatusCode != 200 {
		return []LauncherSchema{}
	}

	responseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return []LauncherSchema{}
	}

	var schemaMapResponse = map[string][]string{}

	if err := json.Unmarshal(responseBody, &schemaMapResponse); err != nil {
		log.Print(err)
		return []LauncherSchema{}
	}

	for surveyType, schemas := range schemaMapResponse {
		for _, schemaName := range schemas {
			schemaList = append(schemaList, LauncherSchemaFromFilename(schemaName, surveyType))
		}
	}

	return schemaList
}

// FindSurveyByName Finds the schema in the list of available schemas
func FindSurveyByName(name string) LauncherSchema {
	availableSchemas := GetAvailableSchemas()

	for _, schemasBySurveyType := range availableSchemas {
		for _, schema := range schemasBySurveyType {
			if schema.Name == name {
				return schema
			}
		}
	}

	panic("Schema not found")
}

func GetSupplementaryDataSets(surveyId string, periodId string) ([]DatasetMetadata, error) {
	datasetList := []DatasetMetadata{}
	hostURL := settings.Get("SDS_API_BASE_URL")

	client := clients.GetHTTPClient()
	tokenSource, err := oidc.GenerateIdToken()
	client.Transport = &oauth2.Transport{
		Source: tokenSource,
	}

	if err != nil {
		log.Print(err)
		return datasetList, errors.New("unable to generate authentication credentials")
	}

	log.Printf("SDS API Base URL: %s", hostURL)
	url := fmt.Sprintf("%s/v1/dataset_metadata?survey_id=%s&period_id=%s", hostURL, surveyId, periodId)
	resp, err := client.Get(url)

	if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 404) {
		return datasetList, errors.New("unable to fetch supplementary data")
	}
	if resp.StatusCode == 404 {
		return datasetList, nil
	}
	responseBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return datasetList, errors.New("unable to read response body of supplementary data")
	}

	if err := json.Unmarshal(responseBody, &datasetList); err != nil {
		log.Print(err)
		return datasetList, fmt.Errorf("%v", err)
	}
	return datasetList, nil
}

// Return a LauncherSchema instance by loading schema from name or URL
func GetLauncherSchema(schemaName string, schemaUrl string) LauncherSchema {
	var launcherSchema LauncherSchema

	if schemaUrl != "" {
		log.Println("Getting schema by URL: " + schemaUrl)
		launcherSchema = LauncherSchema{
			URL:  schemaUrl,
			Name: schemaName,
		}
	} else if schemaName != "" {
		log.Println("Searching for schema by name: " + schemaName)
		launcherSchema = FindSurveyByName(schemaName)
	} else {
		panic("Either `schema_name` or `schema_url` must be provided.")
	}

	return launcherSchema
}
