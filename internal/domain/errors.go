package domain

import "fmt"

type ConnectionError struct {
	Service string
	Err     error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection to %s failed: %v", e.Service, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

type ServiceHealth struct {
	S3      bool
	SQS     bool
	SNS     bool
	Secrets bool
}

func (h ServiceHealth) AllOK() bool {
	return h.S3 && h.SQS && h.SNS && h.Secrets
}
