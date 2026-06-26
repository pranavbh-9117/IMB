package workerpool

// Metrics holds runtime observability statistics for the worker pool.
type Metrics struct {
	WorkersCount  int   // Total configured worker goroutines
	ActiveWorkers int64 // Currently executing workers
	QueuedJobs    int   // Jobs currently waiting in queue buffer
	CompletedJobs int64 // Total jobs executed successfully
	FailedJobs    int64 // Total jobs that returned errors or panicked
}
