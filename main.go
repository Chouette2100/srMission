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
000100 2026-08-31 csrftokenとcookieを取得してAPIと連携する(コメント投稿とミッション達成状況の確認)
000200 2026-09-02 広告視聴のためのviewReawrd()を追加する(まだ意図したとおりに動作しない)
000300 2026-09-05 000200/viewReward.goのレビュー、修正を行う（iFrame対応）
000301 2026-09-06 広告視聴を5回繰り返しても報酬が得られないときは待ち時間を大幅に増やすようにする
000302 2026-09-06 環境変数SR_TRACEBACKを有効にすると、SIGINT,SIGTERM時にgoroutineのスタックトレースを出力するようにする
                  mission == dailyで有効な視聴ルーム数が20ルームに達したら処理を打ち切る
*/

const Version = "000302"

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
	// 環境変数SR_TRACEBACKを有効にすると、SIGINT,SIGTERM時にgoroutineのスタックトレースを出力するようにする
	if isTruthyEnv(os.Getenv("SR_TRACEBACK")) {
		log.Printf("SR_TRACEBACK is enabled. SIGINT,SIGTERM will dump goroutine traceback.\n")
		installSignalTracebackHandlers()
	} else {
		log.Printf("SR_TRACEBACK is disabled. SIGINT,SIGTERM will terminate without traceback.\n")
	}

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
	log.Printf("Env.yml  evnConfig.SrAcct = %s\n", envConfig.SrAcct)
	// --------------------------------

	/// 環境変数から設定値を取得する

	var mission, comment string
	var viewingTime int
	// 起動時パラメータからeventid, ibreg, ieregを取得する。
	if len(os.Args) < 5 {
		log.Printf("Usage: srAddEvent eventid ibreg iereg\n")
		return
	}
	mission = os.Args[1]
	noofrooms, _ := strconv.Atoi(os.Args[2])
	viewingTime, _ = strconv.Atoi(os.Args[3])
	comment = os.Args[4]
	log.Printf(" mission =[%s], viewingTime=%d, comment=%s\n", mission, viewingTime, comment)
	// --------------------------------

	// login操作を行う
	defer closeBrowser()
	if err = srLogin(envConfig.SrAcct, envConfig.SrPswd); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	apiClient, apiJar, err := PrepareAPIClientFromCurrentBrowser(envConfig.SrAcct, "https://www.showroom-live.com/")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	defer apiJar.Save()

	csrfToken, err := FetchAPICSRFToken(apiClient)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	log.Printf("API session prepared. csrf_token acquired (length=%d)\n", len(csrfToken))

	switch mission {
	case "daily", "newcommer":
		log.Printf("Mission: %s\n", mission)
		// 視聴の対象となる配信者のURLのリストを取得する
		rooms, err := collectRooms(mission, noofrooms)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}
		// TODO: viewingTimeづつ視聴を行う
		for _, room := range rooms {
			log.Printf("Room: %+v\n", room)
			if err = viewRoom(apiClient, csrfToken, mission, room, viewingTime, comment); err != nil {
				if err.Error() == cmsg {
					break
				}
				log.Printf("Error: %v\n", err)
			}
		}
	case "viewreward":
		log.Printf("Mission: viewreward\n")
		if err = viewReward(apiClient, csrfToken); err != nil {
			log.Printf("Error: %v\n", err)
		}
	default:
		log.Printf("Unknown mission: %s\n", mission)
		return
	}

	// --------------------------------
	log.Printf("End\n")
}
