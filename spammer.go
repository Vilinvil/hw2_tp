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
	// 	in - User
	// 	out - MsgID
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
