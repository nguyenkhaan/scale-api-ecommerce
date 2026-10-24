package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"github.com/joho/godotenv"
)

//Define a struct for config type
type Config struct {
	DatabaseUrl string
	Port string
	Env string
	//string, int
}
func loadDotEnv() error {
	err := godotenv.Load();
	if err == nil || errors.Is(err , os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("Error while loading .env %v", err)
}

func Load() (Config, error) {
	if err := loadDotEnv(); err != nil {
		return Config{} , err
		//Create an instance from struct -> using StructName{ } 
	}
	databaseUrl := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	//using strings package for string format (Trim, Compare, Contains)
	//string là 1 package cung cấp rất nhiều các hàm xử lý chuỗi thật ultra (Compare, Contains, IndexOf...)
	port := strings.TrimSpace(os.Getenv("PORT"))
	if (port == "") {
		port = "4000"
	}
	env := strings.TrimSpace(os.Getenv("ENV"))
	return Config{
		DatabaseUrl: databaseUrl,
		Port: port,
		Env : env,
	}, nil //Khi khai bao bien voi struct thi khong dung "" cho key, "" chi dung cho key, con struct thi thi viet nhu binh thuong vi cac key da duoc co dinh san ten bien roi
	
}
 