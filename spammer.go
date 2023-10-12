package main

import "sync"

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
	// 	in - string
	// 	out - User
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
