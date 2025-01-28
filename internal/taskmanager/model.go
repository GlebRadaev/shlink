// Package taskmanager defines the structure and behavior of tasks that can be managed within a worker pool system.
package taskmanager

// Task represents a task interface that defines the TaskType method.
type Task interface {
	// TaskType returns the type of the task as a string.
	// It is used to distinguish between different task types in the system.
	TaskType() string
}

// DeleteTask represents a task that involves deleting URLs associated with a specific user.
type DeleteTask struct {
	// UserID is the unique identifier of the user who is associated with the URLs to be deleted.
	UserID string

	// URLs is a slice of URLs to be deleted for the user.
	URLs []string
}

// TaskType returns the task type identifier for the DeleteTask.
func (DeleteTask) TaskType() string {
	return "delete_urls_task"
}
