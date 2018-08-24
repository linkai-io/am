package am

type WorkerConfig struct {
	QueueName       string
	CoordinatorAddr string
	Type            ModuleType
}

type WorkerReport struct {
	MessagesProcessed int32
	MessagesInflight  int32
	IOErrors          int32
	AverageTime       int64
}
