package main

import "time"

type Task struct {
	ID       string `json:"id"`
	Contents string `json:"contents"`

	// For in-progress tasks.
	expiration time.Time

	queuePrev *Task
	queueNext *Task
}

func (t *Task) DisconnectedCopy() *Task {
	return &Task{ID: t.ID, Contents: t.Contents}
}

type TaskDeque struct {
	first *Task
	last  *Task
	count int
}

// DecodeTaskDeque inverts TaskDeque.Encode(), converting a serializable
// object back into a linked list deque.
func DecodeTaskDeque(obj []EncodedTask) *TaskDeque {
	res := &TaskDeque{count: len(obj)}
	for i, et := range obj {
		task := &Task{ID: et.ID, Contents: et.Contents, expiration: et.Expiration}
		if i == 0 {
			res.first = task
			res.last = task
		} else {
			res.last.queueNext = task
			task.queuePrev = res.last
			res.last = task
		}
	}
	return res
}

// Encode generates a JSON-serializable object for the task sequence.
// This can be reversed by DecodeTaskDeque.
func (t *TaskDeque) Encode() []EncodedTask {
	objs := make([]EncodedTask, 0, t.count)
	t.Iterate(func(obj *Task) {
		objs = append(objs, EncodedTask{
			ID:         obj.ID,
			Contents:   obj.Contents,
			Expiration: obj.expiration,
		})
	})
	return objs
}

func (t *TaskDeque) Len() int {
	return t.count
}

func (t *TaskDeque) PushLast(task *Task) {
	t.count += 1
	if t.last == nil {
		t.first = task
		t.last = task
		task.queuePrev = nil
		task.queueNext = nil
	} else {
		t.last.queueNext = task
		task.queuePrev = t.last
		task.queueNext = nil
		t.last = task
	}
}

func (t *TaskDeque) PushFirst(task *Task) {
	t.count += 1
	if t.first == nil {
		t.first = task
		t.last = task
		task.queuePrev = nil
		task.queueNext = nil
	} else {
		t.first.queuePrev = task
		task.queueNext = t.last
		task.queuePrev = nil
		t.first = task
	}
}

func (t *TaskDeque) PopLast() *Task {
	res := t.last
	if res != nil {
		t.Remove(res)
	}
	return res
}

func (t *TaskDeque) PopFirst() *Task {
	res := t.first
	if res != nil {
		t.Remove(res)
	}
	return res
}

func (t *TaskDeque) PeekFirst() *Task {
	return t.first
}

func (t *TaskDeque) Remove(task *Task) {
	if task.queuePrev == nil {
		if t.first != task {
			panic("task not in deque")
		}
		t.first = task.queueNext
		task.queueNext = nil
		if t.first != nil {
			t.first.queuePrev = nil
		} else {
			t.last = nil
		}
	} else if task.queueNext == nil {
		if t.last != task {
			panic("task not in queue")
		}
		t.last = task.queuePrev
		task.queuePrev = nil
		if t.last != nil {
			t.last.queueNext = nil
		} else {
			t.first = nil
		}
	} else {
		task.queueNext.queuePrev = task.queuePrev
		task.queuePrev.queueNext = task.queueNext
		task.queueNext = nil
		task.queuePrev = nil
	}
	t.count -= 1
	if task.queueNext != nil || task.queuePrev != nil {
		panic("pointer unexpectedly preserved")
	}
}

func (t *TaskDeque) Iterate(f func(t *Task)) {
	obj := t.first
	for obj != nil {
		f(obj)
		obj = obj.queueNext
	}
}

type EncodedTask struct {
	ID         string
	Contents   string
	Expiration time.Time
}
