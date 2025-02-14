/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.14.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

type ModelEndpointAlert struct {
	ModelId         int32                         `json:"model_id,omitempty"`
	ModelEndpointId int32                         `json:"model_endpoint_id,omitempty"`
	EnvironmentName string                        `json:"environment_name,omitempty"`
	TeamName        string                        `json:"team_name,omitempty"`
	AlertConditions []ModelEndpointAlertCondition `json:"alert_conditions,omitempty"`
}
