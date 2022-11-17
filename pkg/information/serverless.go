package information

import (
	"os"
)

func isAWSLambda() bool {
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); !ok {
		return false
	}

	return true
}

func isGoogleCloudFunction() bool {
	if _, ok := os.LookupEnv("FUNCTION_TARGET"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("FUNCTION_SIGNATURE_TYPE"); !ok {
		return false
	}

	return true
}

func isCloudRunOrFireBase() bool {
	if _, ok := os.LookupEnv("K_SERVICE"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("K_REVISION"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("K_CONFIGURATION"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("PORT"); !ok {
		return false
	}

	return true
}

func isAzureFunction() bool {
	if _, ok := os.LookupEnv("FUNCTIONS_WORKER_RUNTIME"); !ok {
		return false
	}

	if _, ok := os.LookupEnv("WEBSITE_SITE_NAME"); !ok {
		return false
	}

	return true
}

func addKeyToLabels(key string, value string, info *AgentInformation) {
	if info.Labels == nil {
		info.Labels = make(map[string]string)
	}

	if value == "" {
		return
	}

	if _, ok := info.Labels[key]; !ok {
		info.Labels[key] = value
	}
}

func collectServerless(info *AgentInformation) error {
	var functionName string

	if isAWSLambda() {
		functionName, _ = os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
		awsRegion, _ := os.LookupEnv("AWS_REGION")
		addKeyToLabels("aws_region", awsRegion, info)
	} else if isGoogleCloudFunction() || isCloudRunOrFireBase() {
		var ok bool
		functionName, ok = os.LookupEnv("FUNCTION_NAME")
		if !ok {
			functionName, _ = os.LookupEnv("K_SERVICE")
		}
	} else if isAzureFunction() {
		functionName, _ = os.LookupEnv("WEBSITE_SITE_NAME")
		azureRegion, _ := os.LookupEnv("REGION_NAME")
		addKeyToLabels("azure_region", azureRegion, info)
	}

	if functionName != "" {
		addKeyToLabels("rookout_serverless", "true", info)
		addKeyToLabels("function_name", functionName, info)
	}

	return nil
}
