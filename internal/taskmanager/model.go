package taskmanager

type Task interface {
	TaskType() string
}

type DeleteTask struct {
	UserID string
	URLs   []string
}

func (DeleteTask) TaskType() string {
	return "delete_urls_task"
}
