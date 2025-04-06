package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	User1IdToken string
	User2IdToken string
	ApiUrl       string
}

func newConfig() config {
	var config config

	// List of env files to load
	envFiles := []string{
		"../../configs/e2e/e2e.env",
		"../../configs/aws/apigateway.env",
	}

	// Load all env files
	err := loadEnvFiles(envFiles)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	config.User1IdToken = viper.GetString("USER_1_ID_TOKEN")
	config.User2IdToken = viper.GetString("USER_2_ID_TOKEN")
	config.ApiUrl = viper.GetString("API_URL")

	return config
}

func loadEnvFiles(filenames []string) error {
	for _, file := range filenames {
		viper.SetConfigFile(file) // Set specific file
		viper.SetConfigType("env")
		viper.AutomaticEnv() // Allow override by OS environment variables

		err := viper.MergeInConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

// prettyPrintJSON formats JSON for readability
func prettyPrintJSON(data string) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(data), "", "  ")
	if err != nil {
		return data // Return unformatted if error
	}
	return prettyJSON.String()
}

// Log API request/response to a Markdown file
func logToMarkdown(method, url string, reqHeaders map[string]string, reqBody string, respHeaders map[string][]string, respBody string, statusCode int) error {
	file, err := os.OpenFile("api_logs.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format headers
	formatHeaders := func(headers map[string][]string) string {
		result := ""
		for key, values := range headers {
			result += fmt.Sprintf("**%s:** %s  \n", key, values)
		}
		return result
	}

	logEntry := fmt.Sprintf(
		"## API Test - %s\n\n"+
			"### Request\n"+
			"**Method:** %s  \n"+
			"**URL:** %s  \n\n"+
			"**Headers:**  \n%s\n\n"+
			"**Body:**\n```json\n%s\n```\n\n"+
			"### Response\n"+
			"**Status Code:** %d  \n\n"+
			"**Headers:**  \n%s\n\n"+
			"**Body:**\n```json\n%s\n```\n\n---\n",
		timestamp, method, url,
		formatHeaders(map[string][]string{ // Convert request headers
			"Content-Type": {reqHeaders["Content-Type"]},
		}),
		prettyPrintJSON(reqBody),
		statusCode,
		formatHeaders(respHeaders),
		prettyPrintJSON(respBody),
	)

	_, err = file.WriteString(logEntry)
	return err
}
