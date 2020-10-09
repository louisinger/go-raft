package internal

import ( 
	"math" 
	"log"
 )

type StateMachine struct {
	state Store
	commitIndex int
	lastApplied int
	log []Command
	replicate chan bool
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		state: make(map[string][]byte),
		commitIndex: 0,
		lastApplied: 0,
		log: []Command{},
		replicate: make(chan bool),
	}
}

func (sm StateMachine) getLastLogTerm() int {
	size := len(sm.log)
	if (size == 0) {
		return 0
	}
	return sm.log[size-1].getTerm()
}

func (sm *StateMachine) applyUntilCommitIndex() error {
	nbToApply := sm.commitIndex - sm.lastApplied
	if nbToApply <= 0 {
		sm.replicate <- true
		return nil
	} 
	toApply := sm.log[sm.lastApplied]
	sm.lastApplied++
	log.Println("Apply command to state machine:", toApply)
	err := sm.state.Apply(toApply)
	if err != nil {
		return err
	}
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
	lastIndex := len(sm.log) - 1
	if lastIndex < startIndex {
		return false
	}

	toCheck := sm.log[startIndex:]
	if len(toCheck) == 0 {
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