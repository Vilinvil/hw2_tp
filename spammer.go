package main

import (
	"fmt"
	"sync"

	"github.com/Vilinvil/hw2_tp/pkg/async"
)

func RunPipeline(cmds ...cmd) {
	chIn := make(chan any)
	chOut := make(chan any)
	wg := &sync.WaitGroup{}

	for _, cmdCur := range cmds {
		wg.Add(1)

		chInGo := chIn
		chOutGo := chOut
		cmdCur := cmdCur

		go func() {
			cmdCur(chInGo, chOutGo)
			close(chOutGo)
			wg.Done()
		}()

		chIn = chOutGo
		chOut = make(chan any)
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mapUserUniq := async.New[uint64, struct{}](0)

	for curEmail := range in {
		wg.Add(1)

		curEmail := curEmail
		go func() {
			emailStr, ok := curEmail.(string)
			if !ok {
				fmt.Printf("error cast to string curEmail: %v\n", curEmail)

				return
			}

			user := GetUser(emailStr)
			if _, ok = mapUserUniq.Get(user.ID); !ok {
				mapUserUniq.Insert(user.ID, struct{}{})
				out <- user
			}
		}()
	}

	wg.Done()
}

func SelectMessages(in, out chan interface{}) {
	chBatch := make(chan []User)
	wgGoroutineStart := &sync.WaitGroup{}
	wgGoroutineStart.Add(1)

	go func() {
		wgGoroutineStart.Done()
		for {
			batchUsers := make([]User, 0, GetMessagesMaxUsersBatch)
			for idxUser := 0; idxUser < len(batchUsers); idxUser++ {
				curUser, ok := <-in
				if !ok {
					chBatch <- batchUsers
					close(chBatch)

					return
				}

				user, ok := curUser.(User)
				if !ok {
					fmt.Printf("error cast to user curUser: %v\n", curUser)

					return
				}

				batchUsers[idxUser] = user
			}

			chBatch <- batchUsers
		}
	}()
	wgGoroutineStart.Wait()

	for batch := range chBatch {
		slMsgID, err := GetMessages(batch...)
		if err != nil {
			fmt.Printf("error in GetMessages: %s\n", err.Error())

			return
		}

		for _, msgID := range slMsgID {
			out <- msgID
		}
	}
}

func CheckSpam(in, out chan interface{}) {
	wgWorkersStart := &sync.WaitGroup{}
	chToHash := make(chan MsgID)

	for i := 0; i < HasSpamMaxAsyncRequests; i++ {
		wgWorkersStart.Add(1)

		go func() {
			wgWorkersStart.Done()

			for msgIDToHash := range chToHash {
				hashSpam, err := HasSpam(msgIDToHash)
				if err != nil {
					fmt.Printf("error in HasSpam: %s\n", err.Error())

					return
				}

				out <- MsgData{ID: msgIDToHash, HasSpam: hashSpam}
			}
		}()
	}

	wgWorkersStart.Wait()

	for curMsgID := range in {
		msgID, ok := curMsgID.(MsgID)
		if !ok {
			fmt.Printf("error cast to MsgId curMsgID: %v\n", curMsgID)

			return
		}

		chToHash <- msgID
	}
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
