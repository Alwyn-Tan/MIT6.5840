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
	LastTaskType  string
	LastTaskIndex int
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

//TaskId: MAP_0 REDUCE_0
//tempImmediateFile: temp-immed-12345-0-2
//finalImmediateFile: immed-0-2
//tempOutputFile: temp-out-2
//finalOutputFile: mr-out-2

func generateTaskId(taskType string, taskIndex int) string {
	return taskType + "_" + strconv.Itoa(taskIndex)
}

func tempImmediateFileName(workerId string, mapIndex int, reduceIndex int) string {
	return "temp-immed-" + workerId + "-" + strconv.Itoa(mapIndex) + "-" + strconv.Itoa(reduceIndex) + ".txt"
}

func finalImmediateFileName(mapIndex int, reduceIndex int) string {
	return "immed-" + strconv.Itoa(mapIndex) + "-" + strconv.Itoa(reduceIndex) + ".txt"
}

func tempOutputFileName(workerId string, reduceIndex int) string {
	return "temp-out-" + strconv.Itoa(reduceIndex) + ".txt"
}

func finalOutputFileName(reduceIndex int) string {
	return "mr-out-" + strconv.Itoa(reduceIndex) + ".txt"
}
