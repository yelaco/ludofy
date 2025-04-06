package entities

type ApplicationEndpoint struct {
	UserId      string `dynamodbav:"UserId"`
	DeviceToken string `dynamodbav:"DeviceToken"`
	EndpointArn string `dynamodbav:"EndpointArn"`
}
