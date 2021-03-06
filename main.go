package main

import (
	"fmt"

	"io/ioutil"

	"flag"

	"github.com/go-yaml/yaml"
	"github.com/iost-official/luckybet-backend/database"
	"github.com/iost-official/luckybet-backend/handler"
	"github.com/iost-official/luckybet-backend/nonce"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttprouter"
	"gopkg.in/mgo.v2"
)

var router fasthttprouter.Router

type Config struct {
	Main struct {
		Watch bool
		Nonce string
	}
	Blockchain struct {
		Contract string
		Server   string
	}
	Database struct {
		Server string
		Name   string
	}
}

func main() {

	var cf string

	cf = "config.yml"

	isWatch := flag.Bool("w", false, "watch of block chain")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(cf)

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	config.Main.Watch = *isWatch
	fmt.Println(config)

	if err != nil {
		panic(err)
	}

	session, err := mgo.Dial(config.Database.Server)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	database.Contract = config.Blockchain.Contract
	database.LocalIServer = config.Blockchain.Server

	err = session.DB("lucky_bet").C("bets").EnsureIndexKey("account", "nonce", "bettime")
	if err != nil {
		fmt.Println(err)
	}
	err = session.DB("lucky_bet").C("results").EnsureIndexKey("time", "round")
	if err != nil {
		fmt.Println(err)
	}
	err = session.DB("lucky_bet").C("rewards").EnsureIndexKey("round", "account")
	if err != nil {
		fmt.Println(err)
	}
	err = session.DB("lucky_bet").C("blocks").EnsureIndexKey("height")
	if err != nil {
		fmt.Println(err)
	}

	handler.D = database.NewDatabase(session.DB(config.Database.Name))

	if config.Main.Watch {
		go handler.D.Watch()
	}

	nonce.D = handler.D
	handler.NonceUrl = config.Main.Nonce

	run()
}

func run() {
	router.POST("/api/luckyBet", handler.LuckyBet)
	router.POST("/api/luckyBetBenchMark", handler.LuckyBetBenchMark)
	router.GET("/api/luckyBet/round/:id", handler.BetRound)
	router.GET("/api/luckyBet/addressBet/:id", handler.AddressBet)
	router.GET("/api/luckyBet/latestBetInfo", handler.LatestBetInfo)
	router.GET("/api/luckyBet/todayRanking", handler.TodayTop10Address)
	router.GET("/api/luckyBetBlockInfo", handler.BetInfo)
	router.GET("/nonce", nonce.Handler)

	err := fasthttp.ListenAndServe(":12345", router.Handler)
	if err != nil {
		panic(err)
	}
}
