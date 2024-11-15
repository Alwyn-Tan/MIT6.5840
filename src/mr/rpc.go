package mr

import (
	"os"
)
import "strconv"

type Task struct {
	Index    int
	TaskType string
	FileName string
}

type ApplyForTaskArgs struct {
	WorkerId int
}

type ApplyForTaskReply struct {
	Task      Task
	ReduceNum int
}

func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
