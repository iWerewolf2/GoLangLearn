package main

import (
	"fmt"
	"sync"
)

type Todo struct {
	Id   int
	Name string
}

type Todos []Todo

type TodoRepo struct {
	mu        sync.Mutex
	currentId int
	todos     Todos
}

// NewTodoRepo initializes the repo with seed data
func NewTodoRepo() *TodoRepo {
	repo := &TodoRepo{}
	repo.CreateTodo(Todo{Name: "Write presentation"})
	repo.CreateTodo(Todo{Name: "Host meetup"})
	return repo
}

func (r *TodoRepo) FindTodo(id int) (Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.todos {
		if t.Id == id {
			return t, nil
		}
	}
	return Todo{}, fmt.Errorf("Todo with id %d not found", id)
}

func (r *TodoRepo) CreateTodo(t Todo) Todo {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.currentId++
	t.Id = r.currentId
	r.todos = append(r.todos, t)
	return t
}

func (r *TodoRepo) DestroyTodo(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.todos {
		if t.Id == id {
			r.todos = append(r.todos[:i], r.todos[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Could not find Todo with id %d to delete", id)
}
