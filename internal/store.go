package internal

import (
	"encoding/json"
	"errors"
)

type Json map[string]interface{}
type Store map[string][]byte

type Command interface {
	apply(state *Store) error
	getTerm() int
}

type Entry struct {
	Term int
}

// The command put inserts (or replaces) a value at a given key
type Put struct {
	Entry
	key string
	value Json
}

// The command delete removes a key from the given state.
type Delete struct {
	Entry
	key string
}

func (e Entry) getTerm() int {
	return e.Term
}

// apply a list of commands (= a log)
func (state *Store) Apply(cmd Command) error {
	if err := cmd.apply(state); err != nil {
		return err
	}
	return nil
}

// method using to get the value of a specific key
func (state *Store) Get(key string) (Json, error) {
	bytes, ok := (*state)[key]
	if (ok == false) {
		return nil, errors.New("Can't get the value for the key: " + key)
	}

	var jsonResult Json

	if err := json.Unmarshal(bytes, &jsonResult); err != nil {
		return nil, err
	}

	return jsonResult, nil
}


// implements apply function for PUT and DELETE

func (put Put) apply(state *Store) error {
	bytes, err := json.Marshal(put.value)
	if err != nil {
		return err
	}
	(*state)[put.key] = bytes
	return nil
}

func (del Delete) apply(state *Store) error {
	if _, ok := (*state)[del.key]; ok == false {
		return errors.New("The key: " + del.key + " does not exist.")
	}
	delete(*state, del.key)
	return nil
}


