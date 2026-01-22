package main

import (
	"fmt"
	"my_crossstich_shop/pkg/config"
	"my_crossstich_shop/pkg/repository"
	"my_crossstich_shop/pkg/server"
)

var db repository.DB

func main() {
	fmt.Println("🚀 Запуск сервера...")

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	db, err = repository.New(cfg)
	if err != nil {
		panic(err)
	}

	srv := server.New(db)
	err = srv.Run()
	if err != nil {
		panic(err)
	}
}
