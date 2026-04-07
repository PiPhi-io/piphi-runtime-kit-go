package runtimekit

// TrackBackgroundTask runs one goroutine while updating the tracked task count.
func TrackBackgroundTask(processState *RuntimeProcessState, fn func()) {
	if processState == nil || fn == nil {
		return
	}

	processState.incrementPendingTasks()
	go func() {
		defer processState.decrementPendingTasks()
		fn()
	}()
}

// CreateTrackedTask is an alias to match the Python and Node kit naming.
func CreateTrackedTask(processState *RuntimeProcessState, fn func()) {
	TrackBackgroundTask(processState, fn)
}

// WaitPendingBackgroundTasks blocks until tracked goroutines settle.
func WaitPendingBackgroundTasks(processState *RuntimeProcessState) {
	if processState == nil {
		return
	}
	processState.WaitPendingTasks()
}
