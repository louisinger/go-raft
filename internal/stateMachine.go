package internal

type StateMachine struct {
	State Store
	Log []Command
}

// func (stateMachine *StateMachine) 