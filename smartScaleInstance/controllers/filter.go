package controllers

import (
	"fmt"
	"github.com/astaxie/beego/context"
	"net/http"
	"smartScaleInstance/token"
	"strings"
)

var FilterToken = func(ctx *context.Context) {
	a := ctx.Request.Header.Get("Authorization")
	if a == "" {
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/views/login.html", http.StatusMovedPermanently)
		return
	}

	tokenarray := strings.Split(a, " ")
	t := ""
	if len(tokenarray) == 2 {
		if tokenarray[0] != "bearer" && tokenarray[0] != "Bearer" {
			http.Redirect(ctx.ResponseWriter, ctx.Request, "/views/login.html", http.StatusMovedPermanently)
			return
		} else {
			t = tokenarray[1]
		}
	} else {
		t = tokenarray[0]
	}

	if err := token.CheckToken(t); err != nil {
		fmt.Printf("token is invalid.\n")
		http.Redirect(ctx.ResponseWriter, ctx.Request, "/views/login.html", http.StatusMovedPermanently)
		return
	}
	fmt.Printf("token check ok.\n")

}
