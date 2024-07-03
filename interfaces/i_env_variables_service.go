package interfaces

import "job_sender/types"

type IEnvVariablesService interface {
	// Get env variables from the cloudbuild.yaml file
	GetEnvVariables() *types.EnvVariables
}
