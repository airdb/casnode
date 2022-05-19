// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discuzx

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/casbin/casnode/casdoor"
	"github.com/casbin/casnode/controllers"
	"github.com/casbin/casnode/object"
	"github.com/casdoor/casdoor-go-sdk/auth"
	"github.com/go-sql-driver/mysql"
)

var AddUsersConcurrency = 20

func TestAddUsers(t *testing.T) {
	object.InitConfig()
	InitAdapter()
	object.InitAdapter()
	casdoor.InitCasdoorAdapter()
	controllers.InitAuthConfig()

	membersEx := getMembersEx()

	var wg sync.WaitGroup
	wg.Add(len(membersEx))

	sem := make(chan int, AddUsersConcurrency)
	users := []*auth.User{}
	for i, memberEx := range membersEx {
		sem <- 1
		go func(i int, memberEx *MemberEx) {
			defer wg.Done()

			user := getUserFromMember(memberEx)
			users = append(users, user)
			fmt.Printf("[%d/%d]: Added user: [%d, %s]\n", i+1, len(membersEx), memberEx.Member.Uid, memberEx.Member.Username)
			<-sem
		}(i, memberEx)
	}

	wg.Wait()

	casdoor.AddUsersInBatch(users)
}

func TestAddUcenterUsers(t *testing.T) {
	object.InitConfig()
	InitAdapter()
	object.InitAdapter()
	casdoor.InitCasdoorAdapter()
	controllers.InitAuthConfig()

	membersEx := getUcenterMembersEx()

	var wg sync.WaitGroup
	wg.Add(len(membersEx))

	sem := make(chan int, AddUsersConcurrency)
	users := []*auth.User{}
	for i, memberEx := range membersEx {
		sem <- 1
		go func(i int, memberEx *MemberEx) {
			defer wg.Done()
			user := getUserFromUcenterMember(memberEx)
			users = append(users, user)
			fmt.Printf("[%d/%d]: Added user: [%d, %s]\n", i+1, len(membersEx), memberEx.UcenterMember.Uid, memberEx.UcenterMember.Username)
			<-sem
		}(i, memberEx)
	}

	wg.Wait()

	addUsersInBatchWithPanic(users)
}

func addUsersInBatchWithPanic(users []*auth.User) bool {
	batchSize := 1000

	if len(users) == 0 {
		return false
	}

	affected := false
	for i := 0; i < (len(users)-1)/batchSize+1; i++ {
		start := i * batchSize
		end := (i + 1) * batchSize
		if end > len(users) {
			end = len(users)
		}

		tmp := users[start:end]
		fmt.Printf("Add users: [%d - %d].\n", start, end)
		if addUsersWitPanic(tmp) {
			affected = true
		}
	}

	return affected
}

func addUsersWitPanic(users []*auth.User) bool {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
			for i := 0; i < len(users); i++ {
				addUserWitPanic(users[i])
			}
		} else {
			log.Fatal(err)
		}
	}()

	return casdoor.AddUsers(users)
}

func addUserWitPanic(user *auth.User) bool {
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 {
			log.Println(err)
		} else {
			log.Fatal(err)
		}
	}()

	return casdoor.AddUser(user)
}
