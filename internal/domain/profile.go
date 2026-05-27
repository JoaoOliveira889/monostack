package domain

type S3ObjectSnapshot struct {
	Key           string `yaml:"key"`
	Size          int64  `yaml:"size,omitempty"`
	LastModified  string `yaml:"last_modified,omitempty"`
	ContentBase64 string `yaml:"content_base64,omitempty"`
	ContentType   string `yaml:"content_type,omitempty"`
}

type S3BucketSnapshot struct {
	Name    string             `yaml:"name"`
	Objects []S3ObjectSnapshot `yaml:"objects,omitempty"`
}

type SQSQueueSnapshot struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"url,omitempty"`
	ARN        string            `yaml:"arn,omitempty"`
	Attributes map[string]string `yaml:"attributes,omitempty"`
}

type SNSTopicSnapshot struct {
	Name          string            `yaml:"name"`
	ARN           string            `yaml:"arn,omitempty"`
	Subscriptions []SNSSubscription `yaml:"subscriptions,omitempty"`
}

type AppProfile struct {
	Version       int                   `yaml:"version"`
	Config        *AWSConfig            `yaml:"config"`
	Subscriptions []ManagedSubscription `yaml:"subscriptions,omitempty"`
	YamlScript    string                `yaml:"yaml_script,omitempty"`
	S3            []S3BucketSnapshot    `yaml:"s3,omitempty"`
	SQS           []SQSQueueSnapshot    `yaml:"sqs,omitempty"`
	SNS           []SNSTopicSnapshot    `yaml:"sns,omitempty"`
}
