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

package controllers

import (
	beego "github.com/beego/beego/v2/adapter"
	"github.com/casbin/casnode/object"
	"github.com/casbin/casnode/util"
	"github.com/casdoor/casdoor-go-sdk/auth"
)

var CasdoorEndpoint = beego.AppConfig.String("casdoorEndpoint")
var ClientId = beego.AppConfig.String("clientId")
var ClientSecret = beego.AppConfig.String("clientSecret")
var JwtSecret = beego.AppConfig.String("jwtSecret")
var CasdoorOrganization = beego.AppConfig.String("casdoorOrganization")

func init() {
	auth.InitConfig(CasdoorEndpoint, ClientId, ClientSecret, JwtSecret, CasdoorOrganization)
}

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Data2  interface{} `json:"data2"`
}

// @Title Signin
// @Description sign in as a member
// @Param   code     QueryString    string  true        "The code to sign in"
// @Param   state     QueryString    string  true        "The state"
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /signin [post]
func (c *ApiController) Signin() {
	code := c.Input().Get("code")
	state := c.Input().Get("state")

	token, err := auth.GetOAuthToken(code, state)
	if err != nil {
		panic(err)
	}

	claims, err := auth.ParseJwtToken(token.AccessToken)
	if err != nil {
		panic(err)
	}

	username := claims.Name

	affected, err := object.UpdateMemberOnlineStatus(username, true, util.GetCurrentTime())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	claims.AccessToken = token.AccessToken
	c.SetSessionUser(claims)

	c.ResponseOk(claims, affected)
}

// @Title Signout
// @Description sign out the current member
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /signout [post]
func (c *ApiController) Signout() {
	username := c.GetSessionUsername()
	if username != "" {
		_, err := object.UpdateMemberOnlineStatus(username, false, util.GetCurrentTime())
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
	}

	c.SetSessionUser(nil)

	c.ResponseOk()
}

// @Title GetAccount
// @Description Get current account
// @Success 200 {object} controllers.api_controller.Response The Response object
// @router /get-account [get]
func (c *ApiController) GetAccount() {
	if c.RequireSignedIn() {
		return
	}

	claims := c.GetSessionUser()

	c.ResponseOk(claims)
}

func (c *ApiController) UpdateAccountBalance(balance int) {
	claims := c.GetSessionUser()
	claims.Score = balance
	c.SetSessionUser(claims)
}
