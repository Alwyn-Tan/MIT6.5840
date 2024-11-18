package mr

import (
	"os"
	"time"
)
import "strconv"

type Task struct {
	Index    int
	TaskType string
	FileName string
	WorkerId string
	Deadline time.Time
}

type ApplyForTaskArgs struct {
	WorkerId      string
	LastTaskIndex int
	LastTaskType  string
}

type ApplyForTaskReply struct {
	Task      Task
	ReduceNum int
	MapNum    int
}

func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
