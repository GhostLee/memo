package main

import (
	"context"
	"fmt"
	"github.com/GhostLee/memo/leosys/library"
	"github.com/tencentyun/scf-go-lib/cloudevents/scf"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)


func verify(ctx context.Context, event scf.APIGatewayProxyRequest) (bool, error) {
	if event.QueryString["username"]=="" || event.QueryString["password"]=="" {
		return false, nil
	}
	u := library.NewLibUser(event.QueryString["username"], event.QueryString["password"])
	_, err := u.Login()
	if err != nil{
		fmt.Println(err)
		return false, nil
	}
	return true, nil
}

func main()  {
	cloudfunction.Start(verify)
}
