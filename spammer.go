package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/Vilinvil/hw2_tp/pkg/async"
)

func (m *MsgData) toString() string {
	return fmt.Sprintf("%t %d", m.HasSpam, m.ID)
}

func RunPipeline(cmds ...cmd) {
	chIn := make(chan any)
	chOut := make(chan any)
	wg := &sync.WaitGroup{}

	for _, cmdCur := range cmds {
		wg.Add(1)

		chInGo := chIn
		chOutGo := chOut

		go func(cmdCur cmd) {
			cmdCur(chInGo, chOutGo)
			close(chOutGo)
			wg.Done()
		}(cmdCur)

		chIn = chOutGo
		chOut = make(chan any)
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mapUserUniq := async.NewMap[uint64, struct{}](0)

	for curEmail := range in {
		wg.Add(1)

		go func(curEmail any) {
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

			wg.Done()
		}(curEmail)
	}

	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	chBatch := make(chan []User)
	wgGoroutineStart := &sync.WaitGroup{}
	wgGoroutineStart.Add(1)

	go func() {
		wgGoroutineStart.Done()

		for {
			batchUsers := make([]User, GetMessagesMaxUsersBatch)
			for idxUser := 0; idxUser < len(batchUsers); idxUser++ {
				curUser, ok := <-in
				if !ok {
					chBatch <- batchUsers[:idxUser]
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

	wgBatch := &sync.WaitGroup{}

	for batch := range chBatch {
		wgBatch.Add(1)

		go func(batch []User) {
			slMsgID, err := GetMessages(batch...)
			if err != nil {
				fmt.Printf("error in GetMessages: %s\n", err.Error())

				return
			}

			for _, msgID := range slMsgID {
				out <- msgID
			}

			wgBatch.Done()
		}(batch)
	}

	wgBatch.Wait()
}

func CheckSpam(in, out chan interface{}) {
	wgWorkersStart := &sync.WaitGroup{}
	wgHashSpam := &sync.WaitGroup{}
	chToHash := make(chan MsgID)

	for i := 0; i < HasSpamMaxAsyncRequests; i++ {
		wgWorkersStart.Add(1)

		go func() {
			wgWorkersStart.Done()
			wgHashSpam.Add(1)

			for msgIDToHash := range chToHash {
				hashSpam, err := HasSpam(msgIDToHash)
				if err != nil {
					fmt.Printf("error in HasSpam: %s\n", err.Error())

					return
				}

				out <- MsgData{ID: msgIDToHash, HasSpam: hashSpam}
			}

			wgHashSpam.Done()
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

	close(chToHash)

	wgHashSpam.Wait()
}

func CombineResults(in, out chan interface{}) {
	var slTrueMsgData []MsgData

	var slFalseMsgData []MsgData

	for curMsgData := range in {
		msgData, ok := curMsgData.(MsgData)
		if !ok {
			fmt.Printf("error cast to MsgData curMsgData: %v\n", curMsgData)

			return
		}

		if msgData.HasSpam {
			slTrueMsgData = append(slTrueMsgData, msgData)
		} else {
			slFalseMsgData = append(slFalseMsgData, msgData)
		}
	}

	wgSlTrueMsgData := &sync.WaitGroup{}
	wgSlTrueMsgData.Add(1)

	go func() {
		sort.Slice(slTrueMsgData, func(i, j int) bool {
			return slTrueMsgData[i].ID < slTrueMsgData[j].ID
		})

		for _, trueMsgData := range slTrueMsgData {
			out <- trueMsgData.toString()
		}

		wgSlTrueMsgData.Done()
	}()

	sort.Slice(slFalseMsgData, func(i, j int) bool {
		return slFalseMsgData[i].ID < slFalseMsgData[j].ID
	})
	wgSlTrueMsgData.Wait()

	for _, falseMsgData := range slFalseMsgData {
		out <- falseMsgData.toString()
	}
}
