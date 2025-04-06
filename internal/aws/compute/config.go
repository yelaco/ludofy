package compute

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func getTaskMetadata() (TaskMetadata, error) {
	metadataURL := os.Getenv("ECS_CONTAINER_METADATA_URI_V4") + "/task"
	resp, err := http.Get(metadataURL)
	if err != nil {
		return TaskMetadata{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TaskMetadata{}, err
	}

	var metadata TaskMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return TaskMetadata{}, err
	}

	return metadata, nil
}
