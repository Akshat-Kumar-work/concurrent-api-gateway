package main

import (
	"fmt"
	"time"

	"github.com/Akshat-Kumar-work/pvt_go_package/pkg/auth"
)

func main() {
	svc := auth.NewService([]byte("secret"))

	token, err := svc.CreateToken("user-123", time.Hour)
	if err != nil {
		panic(err)
	}

	claims, err := svc.VerifyToken(token)
	if err != nil {
		panic(err)
	}

	fmt.Println("subject:", claims.Subject)
}
