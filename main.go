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
	"strings"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-gorp/gorp"
	"golang.org/x/term"

	"github.com/Chouette2100/exsrapi/v2"
	"github.com/Chouette2100/srapi/v2"
	"github.com/Chouette2100/srdblib/v3"
	"github.com/go-rod/rod/lib/proto"
)

/*
000000 2026-08-30 テストバージョン（ログインと配信者ページの表示）
000100 2026-08-31 csrftokenとcookieを取得してAPIと連携する(コメント投稿とミッション達成状況の確認)
000200 2026-09-02 広告視聴のためのviewReawrd()を追加する(まだ意図したとおりに動作しない)
000300 2026-09-05 000200/viewReward.goのレビュー、修正を行う（iFrame対応）
000301 2026-09-05 広告視聴を5回繰り返しても報酬が得られないときは待ち時間を大幅に増やすようにする
000302 2026-09-05 環境変数SR_TRACEBACKを有効にすると、SIGINT,SIGTERM時にgoroutineのスタックトレースを出力するようにする
                  mission == dailyで有効な視聴ルーム数が20ルームに達したら処理を打ち切る
000303 2026-09-05 リトライが続いたときの待ち時間を5回毎に大幅に増やすようにする
000304 2026-09-06 リトライが続いたときのadRetryCountの再定義を代入に修正する
000305 2026-09-06 mission == "newcommer"のときも、視聴ルーム数が20ルームに達したら処理を打ち切る
000306 2026-09-06 プログレスバーを見失ったらエラーとする。リトライの待ち時間は通常値と5️⃣回に一回の最大値とする
000307 2026-09-06 リトライの待ち時間は通常値と5️⃣回に一回の最大値とする(前回修正は誤り、逆にしていた)
000308 2026-09-08 新人新人ライバー応援キャンペーンの報酬を受け取ることができるようにする(1)
000309 2026-09-09 接続時に表示されるモーダルダイアログを閉じ、不要なダイアログをすべて閉じる。
000310 2026-09-09 viewRoom()のページ生成をmain()のループ外に出し、1ページを使い回すようにする。
000311 2026-09-10 タイミング調整（Sleep()はウェイトのために使い、画面の更新待ちはMustWaitIdle()）を使う）
000312 2026-09-10 ボタン列/リンク列の処理をループで行うとき、一周ごとにボタン列/リンク列の情報を取得する。
000313 2026-09-10 checkReceivedDaily()を導入する。02時台、14時台であればキラキラを受け取る
000314 2026-09-10 APIを使わないコメント投稿を行う。ギフトボックスを表示し位置を変える（ギフトの準備）
000315 2026-09-11 Dailyでギフトが投げられていないときは、星あるいは草を10個投げる。
000316 2026-09-11 getProgressValue()の引数をエレメントからページに変更する(エレメントを長時間使い回さないようにするため)
000317 2026-09-11 コメント投稿をsendComment()にまとめる。throwGift()のセレクター設定の抜けを正す。視聴時刻のカウントをモーダルダイアログを閉じてから始める。
000318 2026-09-11 throwGift()でElementX()でなくElement()を使うように正す。
000319 2026-09-11 コメントボックスを200px上に移動する(コメント入力が画面内で行うため)
000320 2026-09-11 throwGift()で10個投げるときは、ボタン列の最後のボタンを押すようにするなど。
000321 2026-09-14 待ち時間にゆらぎをもたせる
000322 2026-09-15 room-campaignダイアログがあれば、room-campaign-closeボタンを押下する処理を追加
000400 2026-09-16 SW2026に対応する、処理に汎用性をもたせる。
000500 2026-09-16 ログイン情報を保存し、ログイン済みのブラウザを使って処理するようにする。receivableボタンを押下する処理を追加する。
000501 2026-09-17 ミッションの２番目のタブが選択できないなどの問題に対する検討を行う
000502 2026-09-17 ミッション達成とその確認の汎用化を試みる
000503 2026-09-18 密書達成の例外処理を追加する

*/

const Version = "000503"

var Db *sql.DB
var Dbmap *gorp.DbMap

type EnvConfig struct {
	SrAcct string `yaml:"sr_acct"`
	SrPswd string `yaml:"sr_pswd"`
}

// テーブルviewinghistoryに対する構造体
type ViewingHistory struct {
	RoomID   int       `db:"room_id"`
	Mission  string    `db:"mission"`
	ViewedAt time.Time `db:"viewed_at"`
	Valid    bool      `db:"valid"` // 有効なレコードかどうかを示すフラグ
}

var envConfig EnvConfig

// ライブ動画配信サービスにおいて配信予定や配信状況から最適の視聴スケジュールを作成し視聴することを最終目標とする
func main() {

	var err error

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
	tracebackEnabled := isTruthyEnv(os.Getenv("SR_TRACEBACK"))
	// 環境変数SR_TRACEBACKを有効にすると、SIGINT,SIGTERM時にgoroutineのスタックトレースを出力するようにする
	if tracebackEnabled {
		log.Printf("SR_TRACEBACK is enabled. SIGINT,SIGTERM will dump goroutine traceback.\n")
	} else {
		log.Printf("SR_TRACEBACK is disabled. SIGINT,SIGTERM will still trigger graceful shutdown.\n")
	}
	installSignalHandlers(tracebackEnabled)

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
	// Dbmap.AddTableWithName(&ViewingHistory{}, "viewinghistory").SetKeys(false, "valid", "room_id")

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

	// 視聴用のページを作成し、日本語ロケールを適用する
	page, err := srBrowser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		log.Printf("Error: failed to create page: %v\n", err)
		return
	}
	defer page.Close()
	if err = applyJapaneseLocale(page); err != nil {
		log.Printf("Error: failed to apply Japanese locale: %v\n", err)
		return
	}

	var rooms []Room
	switch mission {
	case "daily", "newcommer", "discovery":
		log.Printf("Mission: %s\n", mission)
		// 視聴の対象となる配信者のURLのリストを取得する
		if mission != "discovery" {
			rooms, err = collectRooms(mission, noofrooms)
			if err != nil {
				log.Printf("Error: %v\n", err)
				return
			}
		} else {
			rooms, err = collectMyNextFave(page, noofrooms)
			if err != nil {
				log.Printf("Error: %v\n", err)
				return
			}
		}

		// TODO: viewingTimeづつ視聴を行う
		for _, room := range rooms {
			log.Printf("Room: %+v\n", room)
			// if err = viewRoom(page, apiClient, csrfToken, mission, room, viewingTime, comment); err != nil {
			if err = viewRoom(page, mission, room, viewingTime, comment); err != nil {
				if strings.Contains(err.Error(), cmsg) {
					log.Printf("viewRoom(): Mission completed\n")
					break
				}
				if strings.Contains(err.Error(), " target buttons, but found") {
					// このルームは配信していない
					log.Printf("viewRoom(): this room is not live, skipping to next room\n")
					continue
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
