package internal

import ( "math" )

type StateMachine struct {
	state Store
	commitIndex int
	lastApplied int
	log []Command
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		state: map[string][]byte{},
		commitIndex: 0,
		lastApplied: 0,
		log: []Command{},
	}
}

func (sm StateMachine) getLastLogTerm() int {
	return sm.log[len(sm.log)-1].getTerm()
}

func (sm *StateMachine) applyUntilCommitIndex() error {
	nbToApply := sm.commitIndex - sm.lastApplied
	if nbToApply <= 0 {
		return nil
	} 
	err := sm.state.Apply(sm.log[sm.lastApplied + 1])
	if err != nil {
		return err
	}
	sm.lastApplied++
	return sm.applyUntilCommitIndex()
}

func (sm *StateMachine) setCommitIndex(newCommitIndex int) error {
	sm.commitIndex = int(math.Min(float64(newCommitIndex), float64(len(sm.log) - 1)))
	return sm.applyUntilCommitIndex()
}

func (sm *StateMachine) append(startAt int, cmds ...Command) {
	existings := sm.log[startAt:]
	sm.log = append(sm.log, cmds[len(existings):]...)
}

func (sm *StateMachine) removeIfConflicts(cmds []Command, startIndex int) bool {
	toCheck := sm.log[startIndex:]
	if (len(toCheck) == 0) {
		return false
	}

	var removeAfter int = -1
	for i, cmd := range toCheck {
		if (cmds[i].getTerm() != cmd.getTerm()) {
			removeAfter = i + startIndex
			break
		}
	}

	if (removeAfter >= 0) {
		sm.log = sm.log[:removeAfter]
		return true
	}

	return false
}