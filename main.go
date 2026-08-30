// Copyright © 2025 chouette.21.00@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-gorp/gorp"
	"golang.org/x/term"

	"github.com/Chouette2100/exsrapi/v2"
	"github.com/Chouette2100/srapi/v2"
	"github.com/Chouette2100/srdblib/v3"
)

/*
000000 2026-08-30 テストバージョン（ログインと配信者ページの表示）
*/

const Version = "000000"

var Db *sql.DB
var Dbmap *gorp.DbMap

type EnvConfig struct {
	SrAcct string `yaml:"sr_acct"`
	SrPswd string `yaml:"sr_pswd"`
}

var envConfig EnvConfig

// ライブ動画配信サービスにおいて配信予定や配信状況から最適の視聴スケジュールを作成し視聴することを最終目標とする
func main() {

	// ログファイルの作成
	logfile, err := exsrapi.CreateLogfile(Version, exsrapi.Version, srapi.Version, srdblib.Version)
	if err != nil {
		log.Printf("ログファイルの作成に失敗しました。%v\n", err)
		return
	}
	defer logfile.Close()

	// フォアグラウンド（端末に接続されているか）を判定
	isForeground := term.IsTerminal(int(os.Stdout.Fd()))
	if isForeground {
		// フォアグラウンドならログファイル + コンソール
		log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	} else {
		// バックグラウンドならログファイルのみ
		log.SetOutput(logfile)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Version=%s Start\n", Version)
	// --------------------------------

	// DB接続
	var dbconfig *srdblib.DBConfig
	Db, dbconfig, err = srdblib.OpenDb("DBConfig.enc.yml")
	if err != nil {
		log.Printf("Database error. err = %v\n", err)
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer Db.Close()
	Db.SetMaxOpenConns(8)
	Db.SetMaxIdleConns(12)

	Db.SetConnMaxLifetime(time.Minute * 5)
	Db.SetConnMaxIdleTime(time.Minute * 5)

	defer Db.Close()
	// log.Printf("%+v\n", dbconfig)

	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	Dbmap = &gorp.DbMap{Db: Db,
		Dialect:         dial,
		ExpandSliceArgs: true, //スライス引数展開オプションを有効化する
	}
	// Dbmap.AddTableWithName(&srdblib.User{}, "user").SetKeys(true, "userid")
	// --------------------------------

	// userテーブルの更新判定の閾値、ApiRoomProfile()の実行頻度を設定する
	fileenv := "Env.enc.yml"
	err = exsrapi.LoadConfig(fileenv, &envConfig)
	if err != nil {
		err = fmt.Errorf("exsrapi.Loadconfig(): %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	log.Printf("Env.yml: %+v\n", envConfig)
	// --------------------------------

	/// 環境変数から設定値を取得する

	var mission, comment string
	var viewingTime int
	// 起動時パラメータからeventid, ibreg, ieregを取得する。
	if len(os.Args) < 4 {
		log.Printf("Usage: srAddEvent eventid ibreg iereg\n")
		return
	}
	mission = os.Args[1]
	viewingTime, _ = strconv.Atoi(os.Args[2])
	comment = os.Args[3]
	log.Printf(" mission =[%s], viewingTime=%d, comment=%s\n", mission, viewingTime, comment)
	// --------------------------------

	// login操作を行う
	defer closeBrowser()
	if err = srLogin(envConfig.SrAcct, envConfig.SrPswd); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	// 視聴の対象となる配信者のURLのリストを取得する
	rooms, err := collectRooms(mission)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	// TODO: viewingTimeづつ視聴を行う
	for _, room := range rooms {
		log.Printf("Room: %+v\n", room)
		if err = viewRoom(room.URL, viewingTime, ""); err != nil {
			log.Printf("Error: %v\n", err)
		}
	}

	// --------------------------------
	log.Printf("End\n")
}
