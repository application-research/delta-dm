package util

import "regexp"

// Check if the dataset name is valid
func ValidateDatasetName(datasetName string) bool {

	if len(datasetName) > 254 {
		return false
	}

	// lowercase characters and dashes
	re := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	return re.MatchString(datasetName)
}
