package authentication

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/eq-questionnaire-launcher/clients"
	"github.com/ONSdigital/eq-questionnaire-launcher/settings"
	"github.com/ONSdigital/eq-questionnaire-launcher/surveys"
	"github.com/gofrs/uuid"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/square/go-jose.v2/jwt"

	"bytes"
	"log"
	"path"
	"strconv"
	"strings"
)

// KeyLoadError describes an error that can occur during key loading
type KeyLoadError struct {
	// Op is the operation which caused the error, such as
	// "read", "parse" or "cast".
	Op string

	// Err is a description of the error that occurred during the operation.
	Err string
}

func (e *KeyLoadError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Op + ": " + e.Err
}

// PublicKeyResult is a wrapper for the public key and the kid that identifies it
type PublicKeyResult struct {
	key *rsa.PublicKey
	kid string
}

// PrivateKeyResult is a wrapper for the private key and the kid that identifies it
type PrivateKeyResult struct {
	key *rsa.PrivateKey
	kid string
}

func loadEncryptionKey() (*PublicKeyResult, *KeyLoadError) {
	encryptionKeyPath := settings.Get("JWT_ENCRYPTION_KEY_PATH")

	keyData, err := ioutil.ReadFile(encryptionKeyPath)
	if err != nil {
		return nil, &KeyLoadError{Op: "read", Err: "Failed to read encryption key from file: " + encryptionKeyPath}
	}

	block, _ := pem.Decode(keyData)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, &KeyLoadError{Op: "parse", Err: "Failed to parse encryption key PEM"}
	}

	kid := fmt.Sprintf("%x", sha1.Sum(keyData))

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, &KeyLoadError{Op: "cast", Err: "Failed to cast key to rsa.PublicKey"}
	}

	return &PublicKeyResult{publicKey, kid}, nil
}

func loadSigningKey() (*PrivateKeyResult, *KeyLoadError) {
	signingKeyPath := settings.Get("JWT_SIGNING_KEY_PATH")
	keyData, err := ioutil.ReadFile(signingKeyPath)
	if err != nil {
		return nil, &KeyLoadError{Op: "read", Err: "Failed to read signing key from file: " + signingKeyPath}
	}

	block, _ := pem.Decode(keyData)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, &KeyLoadError{Op: "parse", Err: "Failed to parse signing key from PEM"}
	}

	PublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, &KeyLoadError{Op: "marshal", Err: "Failed to marshal public key"}
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: PublicKey,
	})
	kid := fmt.Sprintf("%x", sha1.Sum(pubBytes))

	return &PrivateKeyResult{privateKey, kid}, nil
}

// QuestionnaireSchema is a minimal representation of a questionnaire schema used for extracting the metadata and questionnaire identifiers
type QuestionnaireSchema struct {
	Metadata   []Metadata `json:"metadata"`
	SchemaName string     `json:"schema_name"`
	SurveyType string     `json:"theme"`
}

// Metadata is a representation of the metadata within the schema with an additional `Default` value
type Metadata struct {
	Name      string `json:"name"`
	Validator string `json:"type"`
	Default   string `json:"default"`
}

func isSurveyMetadata(key string) bool {
	switch key {
	case
		"case_ref",
		"case_type",
		"display_address",
		"employment_date",
		"form_type",
		"period_id",
		"period_str",
		"ref_p_end_date",
		"ref_p_start_date",
		"ru_name",
		"ru_ref",
		"trad_as",
		"user_id",
		"qid",
		"PARTICIPANT_ID",
		"FIRST_NAME",
		"BLOOD_TEST_BARCODE",
		"SWAB_TEST_BARCODE",
		"TEST_QUESTIONS",
		"sds_dataset_id",
		"WINDOW_START_DATE",
		"WINDOW_END_DATE",
		"PORTAL_ID",
		"PARTICIPANT_WINDOW_ID":

		return true
	}
	return false
}

func isRequiredMetadata(key string) bool {
	switch key {
	case
		"case_id",
		"region_code",
		"channel",
		"language_code",
		"collection_exercise_sid",
		"response_expires_at",
		"response_id",
		"schema_name",
		"account_service_url":
		return true
	}
	return false
}

func getSurveyMetadataFromClaims(claimValues map[string][]string, data map[string]interface{}, claims map[string]interface{}, surveyMetadata map[string]interface{}) {
	for key, value := range claimValues {
		if isSurveyMetadata(key) {
			data[key] = value[0]
		} else if key != "roles" && isRequiredMetadata(key) && value[0] != "" {
			claims[key] = value[0]
		} else if key == "roles" {
			claims[key] = value
		}
	}
	surveyMetadata["data"] = data
	claims["survey_metadata"] = surveyMetadata
}

func generateClaims(claimValues map[string][]string, launcherSchema surveys.LauncherSchema) (claims map[string]interface{}) {

	var roles []string
	if rolesValues, ok := claimValues["roles"]; ok {
		roles = rolesValues
	} else {
		roles = []string{"dumper"}
	}

	claims = make(map[string]interface{})

	claims["roles"] = roles
	TxID, _ := uuid.NewV4()
	claims["tx_id"] = TxID.String()

	for key, value := range claimValues {
		if key != "roles" {
			if value[0] != "" {
				claims[key] = value[0]
			}
		} else {
			claims[key] = value
		}
	}
	if len(claimValues["form_type"]) > 0 && len(claimValues["eq_id"]) > 0 {
		log.Println("Deleting schema name from claims")
		delete(claims, "schema_name")
	} else {
		// When quicklaunching, schema_name will not be set, but launcherSchema will have the schema_name.
		if len(claimValues["schema_name"]) == 0 && launcherSchema.Name != "" {
			claims["schema_name"] = launcherSchema.Name
		}
	}

	log.Printf("Using claims: %s", claims)

	return claims
}

func generateClaimsV2(claimValues map[string][]string, launcherSchema surveys.LauncherSchema, schema QuestionnaireSchema) (claims map[string]interface{}) {

	var roles []string
	if rolesValues, ok := claimValues["roles"]; ok {
		roles = rolesValues
	} else {
		roles = []string{"dumper"}
	}

	claims = make(map[string]interface{})

	claims["roles"] = roles
	TxID, _ := uuid.NewV4()
	claims["tx_id"] = TxID.String()
	claims["version"] = "v2"

	surveyMetadata := make(map[string]interface{})
	data := make(map[string]interface{})

	if schema.SurveyType == "social" || schema.SurveyType == "health" {
		receiptingKeys := []string{"qid"}
		surveyMetadata["receipting_keys"] = receiptingKeys
	}

	getSurveyMetadataFromClaims(claimValues, data, claims, surveyMetadata)

	// When quicklaunching, schema_name will not be set, but launcherSchema will have the schema_name.
	if len(claimValues["schema_name"]) == 0 && launcherSchema.Name != "" {
		claims["schema_name"] = launcherSchema.Name
	}

	log.Printf("Using claims: %s", claims)

	return claims
}

// GenerateJwtClaims creates a jwtClaim needed to generate a token
func GenerateJwtClaims() (jwtClaims map[string]interface{}) {
	issued := time.Now()
	expires := issued.Add(time.Minute * 10) // TODO: Support custom exp: r.PostForm.Get("exp")

	jwtClaims = make(map[string]interface{})

	jwtClaims["iat"] = jwt.NewNumericDate(issued)
	jwtClaims["exp"] = jwt.NewNumericDate(expires)
	jti, _ := uuid.NewV4()
	jwtClaims["jti"] = jti.String()

	return jwtClaims
}

func launcherSchemaFromURL(url string) (launcherSchema surveys.LauncherSchema, error string) {
	resp, err := clients.GetHTTPClient().Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		return launcherSchema, fmt.Sprintf("Failed to load Schema from %s", url)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		panic(err)
	}

	validationError := validateSchema(responseBody)
	if validationError != "" {
		return launcherSchema, validationError
	}

	var schema QuestionnaireSchema
	if err := json.Unmarshal(responseBody, &schema); err != nil {
		panic(err)
	}

	cacheBust := ""
	if !strings.Contains(url, "?") {
		cacheBust = "?bust=" + time.Now().Format("20060102150405")
	}

	schemaName := ""

	if schema.SchemaName == "" {
		lastSlash := strings.LastIndex(url, "/")
		if lastSlash != -1 {
			lastDot := strings.LastIndex(url, ".")
			if lastDot == -1 {
				lastDot = len(url)
			}
			schemaName = url[lastSlash+1 : lastDot]
		}
	} else {
		schemaName = schema.SchemaName
	}

	launcherSchema = surveys.LauncherSchema{
		URL:        url + cacheBust,
		Name:       schemaName,
		SurveyType: schema.SurveyType,
	}

	return launcherSchema, ""
}

func validateSchema(payload []byte) (error string) {
	if settings.Get("SCHEMA_VALIDATOR_URL") == "" {
		return ""
	}

	validateURL, _ := url.Parse(settings.Get("SCHEMA_VALIDATOR_URL"))
	validateURL.Path = path.Join(validateURL.Path, "validate")

	log.Println("Validating schema: ", validateURL.String())

	resp, err := http.Post(validateURL.String(), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err.Error()
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err.Error()
	}

	if resp.StatusCode != 200 {
		return string(responseBody)
	}

	return ""
}

func getSchemaClaims(LauncherSchema surveys.LauncherSchema) map[string]interface{} {

	schemaClaims := make(map[string]interface{})
	if LauncherSchema.URL != "" {
		schemaClaims["schema_url"] = LauncherSchema.URL
	}

	return schemaClaims
}

// TokenError describes an error that can occur during JWT generation
type TokenError struct {
	// Err is a description of the error that occurred.
	Desc string

	// From is optionally the original error from which this one was caused.
	From error
}

func (e *TokenError) Error() string {
	if e == nil {
		return "<nil>"
	}
	err := e.Desc
	if e.From != nil {
		err += " (" + e.From.Error() + ")"
	}
	return err
}

// generateTokenFromClaims creates a token though encryption using the private and public keys
func generateTokenFromClaims(cl map[string]interface{}) (string, *TokenError) {
	privateKeyResult, keyErr := loadSigningKey()
	if keyErr != nil {
		return "", &TokenError{Desc: "Error loading signing key", From: keyErr}
	}

	publicKeyResult, keyErr := loadEncryptionKey()
	if keyErr != nil {
		return "", &TokenError{Desc: "Error loading encryption key", From: keyErr}
	}

	opts := jose.SignerOptions{}
	opts.WithType("JWT")
	opts.WithHeader("kid", privateKeyResult.kid)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKeyResult.key}, &opts)
	if err != nil {
		return "", &TokenError{Desc: "Error creating JWT signer", From: err}
	}

	encryptor, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKeyResult.key, KeyID: publicKeyResult.kid},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))

	if err != nil {
		return "", &TokenError{Desc: "Error creating JWT signer", From: err}
	}

	token, err := jwt.SignedAndEncrypted(signer, encryptor).Claims(cl).CompactSerialize()

	if err != nil {
		return "", &TokenError{Desc: "Error signing and encrypting JWT", From: err}
	}

	log.Println("Created signed/encrypted JWT:", token)

	return token, nil
}

func getBooleanOrDefault(key string, values map[string][]string, defaultValue bool) bool {
	if keyValues, ok := values[key]; ok {
		booleanValue, _ := strconv.ParseBool(keyValues[0])
		return booleanValue
	}

	return defaultValue
}

func getStringOrDefault(key string, values map[string][]string, defaultValue string) string {
	if keyValues, ok := values[key]; ok {
		return keyValues[0]
	}

	return defaultValue
}

// GenerateTokenFromDefaults coverts a set of DEFAULT values into a JWT
func GenerateTokenFromDefaults(schemaURL string, accountServiceURL string, accountServiceLogOutURL string, urlValues url.Values) (token string, error string) {
	launcherSchema, validationError := launcherSchemaFromURL(schemaURL)
	if validationError != "" {
		return "", validationError
	}

	claims := make(map[string]interface{})
	urlValues["account_service_url"] = []string{accountServiceURL}
	urlValues["account_service_log_out_url"] = []string{accountServiceLogOutURL}
	claims = generateClaims(urlValues, launcherSchema)

	requiredMetadata, error := GetRequiredMetadata(launcherSchema)
	if error != "" {
		return "", fmt.Sprintf("GetRequiredMetadata failed err: %v", error)
	}

	for _, metadata := range requiredMetadata {
		if metadata.Validator == "boolean" {
			claims[metadata.Name] = getBooleanOrDefault(metadata.Name, urlValues, false)
			continue
		}
		claims[metadata.Name] = getStringOrDefault(metadata.Name, urlValues, metadata.Default)
	}

	jwtClaims := GenerateJwtClaims()
	for key, v := range jwtClaims {
		claims[key] = v
	}

	schemaClaims := getSchemaClaims(launcherSchema)
	for key, v := range schemaClaims {
		claims[key] = v
	}

	token, tokenError := generateTokenFromClaims(claims)
	if tokenError != nil {
		return token, fmt.Sprintf("GenerateTokenFromDefaults failed err: %v", tokenError)
	}

	return token, ""
}

// GenerateTokenFromDefaultsV2 coverts a set of DEFAULT values into a JWT
func GenerateTokenFromDefaultsV2(schemaURL string, accountServiceURL string, urlValues url.Values) (token string, error string) {
	launcherSchema, validationError := launcherSchemaFromURL(schemaURL)
	if validationError != "" {
		return "", validationError
	}

	schema, error := getSchema(launcherSchema)
	if error != "" {
		return "", fmt.Sprintf("getSchema failed err: %v", error)
	}

	claims := make(map[string]interface{})
	urlValues["account_service_url"] = []string{accountServiceURL}

	claims = generateClaimsV2(urlValues, launcherSchema, schema)

	requiredMetadata, error := GetRequiredMetadata(launcherSchema)
	if error != "" {
		return "", fmt.Sprintf("GetRequiredMetadata failed err: %v", error)
	}

	data := make(map[string]interface{})

	for _, metadata := range requiredMetadata {
		if metadata.Validator == "boolean" {
			data[metadata.Name] = getBooleanOrDefault(metadata.Name, urlValues, false)

			continue
		}
		data[metadata.Name] = getStringOrDefault(metadata.Name, urlValues, metadata.Default)
	}

	surveyMetadata := make(map[string]interface{})
	updatedData := make(map[string]interface{})

	if claims["survey_metadata"] != nil {
		surveyMetadata = claims["survey_metadata"].(map[string]interface{})
	}
	for key, value := range data {
		updatedData[key] = value
	}

	initialData := surveyMetadata["data"].(map[string]interface{})

	for key, value := range initialData {
		updatedData[key] = value
	}

	surveyMetadata["data"] = updatedData
	claims["survey_metadata"] = surveyMetadata

	jwtClaims := GenerateJwtClaims()
	for key, v := range jwtClaims {
		claims[key] = v
	}

	schemaClaims := getSchemaClaims(launcherSchema)
	for key, v := range schemaClaims {
		claims[key] = v
	}

	token, tokenError := generateTokenFromClaims(claims)
	if tokenError != nil {
		return token, fmt.Sprintf("GenerateTokenFromDefaults failed err: %v", tokenError)
	}

	return token, ""
}

// TransformSchemaParamsToName Returns a schema name from business schema parameters
func TransformSchemaParamsToName(postValues url.Values) string {
	if postValues.Get("schema_name") != "" {
		return postValues["schema_name"][0]
	}

	eqId := postValues.Get("eq_id")
	formType := postValues.Get("form_type")
	schemaName := fmt.Sprintf("%s_%s", eqId, formType)

	return schemaName
}

// GenerateTokenFromPost converts a set of POST values into a JWT
func GenerateTokenFromPost(postValues url.Values, launchVersion2 bool) (string, string) {
	log.Println("POST received: ", postValues)

	schemaName := TransformSchemaParamsToName(postValues)
	schemaUrl := postValues.Get("schema_url")

	launcherSchema := surveys.GetLauncherSchema(schemaName, schemaUrl)

	schema, error := getSchema(launcherSchema)
	if error != "" {
		return "", fmt.Sprintf("getSchema failed err: %v", error)
	}

	claims := make(map[string]interface{})

	if launchVersion2 {
		claims = generateClaimsV2(postValues, launcherSchema, schema)
	} else {
		claims = generateClaims(postValues, launcherSchema)
	}

	jwtClaims := GenerateJwtClaims()
	for key, v := range jwtClaims {
		claims[key] = v
	}

	schemaClaims := getSchemaClaims(launcherSchema)
	for key, v := range schemaClaims {
		claims[key] = v
	}

	requiredMetadata, error := GetRequiredMetadata(launcherSchema)
	if error != "" {
		return "", fmt.Sprintf("GetRequiredMetadata failed err: %v", error)
	}

	for _, metadata := range requiredMetadata {
		if metadata.Validator == "boolean" {
			_, isset := claims[metadata.Name]
			claims[metadata.Name] = isset
		}
	}

	if launcherSchema.Name != "" && claims["schema_name"] == "" {
		claims["schema_name"] = launcherSchema.Name
	}

	token, tokenError := generateTokenFromClaims(claims)
	if tokenError != nil {
		return token, fmt.Sprintf("GenerateTokenFromPost failed err: %v", tokenError)
	}

	return token, ""
}

// GetRequiredMetadata Gets the required metadata from a schema
func GetRequiredMetadata(launcherSchema surveys.LauncherSchema) ([]Metadata, string) {
	schema, error := getSchema(launcherSchema)
	if error != "" {
		return nil, fmt.Sprintf("getSchema failed err: %v", error)
	}

	defaults := GetDefaultValues()

	for i, value := range schema.Metadata {

		if strings.Contains(value.Name, "BARCODE") {
			schema.Metadata[i].Default = "BAR" + fmt.Sprintf("%08d", rand.Int63n(1e8))
		} else {
			schema.Metadata[i].Default = defaults[value.Name]
		}

		if value.Validator == "boolean" {
			schema.Metadata[i].Default = "false"
		}
	}

	claims := make([]string, 0)
	for _, v := range schema.Metadata {
		claims = append(claims, v.Name)
	}

	mandatoryClaims := getMandatatoryClaims(schema.SurveyType, defaults)

	missingClaims := getMissingMandatoryClaims(claims, mandatoryClaims)

	for _, v := range missingClaims {
		schema.Metadata = append(schema.Metadata, v)
	}

	return schema.Metadata, ""
}

func getSchema(launcherSchema surveys.LauncherSchema) (QuestionnaireSchema, string) {
	var url string

	if launcherSchema.URL != "" {
		url = launcherSchema.URL
	} else {
		hostURL := settings.Get("SURVEY_RUNNER_SCHEMA_URL")

		log.Println("Name: ", launcherSchema.Name)
		url = fmt.Sprintf("%s/schemas/%s", hostURL, launcherSchema.Name)
	}

	log.Println("Loading metadata from schema:", url)

	var schema QuestionnaireSchema
	resp, err := clients.GetHTTPClient().Get(url)
	if err != nil {
		log.Println("Failed to load schema from:", url)
		return schema, fmt.Sprintf("Failed to load Schema from %s", url)
	}

	if resp.StatusCode != 200 {
		log.Print("Invalid response code for schema from: ", url)
		return schema, fmt.Sprintf("Failed to load Schema from %s", url)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Print(err)
		return schema, fmt.Sprintf("Failed to load Schema from %s", url)
	}

	if err := json.Unmarshal(responseBody, &schema); err != nil {
		log.Print(err)
		return schema, fmt.Sprintf("Failed to unmarshal Schema from %s", url)
	}

	return schema, ""
}

func getMandatatoryClaims(surveyType string, defaults map[string]string) []Metadata {
	claims := make([]Metadata, 0)
	if surveyType == "health" || surveyType == "social" {
		claims = []Metadata{
			{"qid", "false", defaults["qid"]},
		}

	} else {
		claims = []Metadata{
			{"ru_ref", "false", defaults["ru_ref"]},
			{"period_id", "false", defaults["period_id"]},
			{"user_id", "false", defaults["user_id"]},
			{"sds_dataset_id", "false", defaults["sds_dataset_id"]},
		}
	}

	return claims
}

func getMissingMandatoryClaims(claims []string, mandatoryClaims []Metadata) []Metadata {
	missingClaims := make([]Metadata, 0)
	for _, v := range mandatoryClaims {
		if !(stringInSlice(v.Name, claims)) {
			missingClaims = append(missingClaims, v)
		}
	}

	return missingClaims
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// GetDefaultValues Returns a map of default values for metadata keys
func GetDefaultValues() map[string]string {
	defaults := make(map[string]string)
	collectionExerciseSid, _ := uuid.NewV4()

	var PARTICIPANT_ID = "ABC-" + fmt.Sprintf("%011d", rand.Int63n(1e11))

	defaults["collection_exercise_sid"] = collectionExerciseSid.String()
	defaults["qid"] = fmt.Sprintf("%016d", rand.Int63n(1e16))
	defaults["version"] = "v2"
	defaults["case_type"] = "B"
	defaults["user_id"] = "UNKNOWN"
	defaults["period_id"] = "201605"
	defaults["period_str"] = "May 2017"
	defaults["collection_exercise_sid"] = collectionExerciseSid.String()
	defaults["ru_ref"] = "12346789012A"
	defaults["ru_name"] = "ESSENTIAL ENTERPRISE LTD."
	defaults["ref_p_start_date"] = "2016-05-01"
	defaults["ref_p_end_date"] = "2016-05-31"
	defaults["return_by"] = "2016-06-12"
	defaults["trad_as"] = "ESSENTIAL ENTERPRISE LTD."
	defaults["employment_date"] = "2016-06-10"
	defaults["region_code"] = "GB-ENG"
	defaults["language_code"] = "en"
	defaults["case_ref"] = "1000000000000001"
	defaults["address_line1"] = "68 Abingdon Road"
	defaults["address_line2"] = ""
	defaults["locality"] = ""
	defaults["town_name"] = "Goathill"
	defaults["postcode"] = "PE12 4GH"
	defaults["display_address"] = "68 Abingdon Road, Goathill"
	defaults["country"] = "E"
	defaults["PARTICIPANT_ID"] = PARTICIPANT_ID
	defaults["FIRST_NAME"] = "John"
	defaults["TEST_QUESTIONS"] = "F"
	defaults["sds_dataset_id"] = "001"
	defaults["WINDOW_START_DATE"] = "2023-03-01"
	defaults["WINDOW_END_DATE"] = "2023-03-31"
	defaults["PORTAL_ID"] = fmt.Sprintf("%07d", rand.Int63n(1e7))
	defaults["PARTICIPANT_WINDOW_ID"] = PARTICIPANT_ID + "-" + fmt.Sprintf("%03d", rand.Int63n(1e3))

	return defaults
}
