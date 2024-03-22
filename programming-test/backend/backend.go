package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
)

// データベース登録情報
// JSONに対応する形式で定義
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Date string `json:"date"`
}

func main() {

	// SQL作成
	db := ConnectDB()
	defer db.Close()

	// テーブルが存在しない場合は新規で作成
	CreateTable(db)

	// フロントエンドからユーザー情報登録のPOSTを貰った時のイベント関数ハンドル
	//http.HandleFunc("/registerInfo", PostHandler) // dbを引数で渡したいので↓の形で対応
	http.HandleFunc("/registerInfo", func(w http.ResponseWriter, r *http.Request) {
		RegisterUserData(w, r, db)
	})

	// フロントエンドからデータベースの読み込み要求を貰った時のイベント関数ハンドル
	http.HandleFunc("/readInfo", func(w http.ResponseWriter, r *http.Request) {
		ReadUserData(w, r, db)
	})

	// フロントエンドからユーザー情報削除の要求を貰った時のイベント関数ハンドル
	http.HandleFunc("/deleteInfo", func(w http.ResponseWriter, r *http.Request) {
		DeleteUserData(w, r, db)
	})

	// サーバー起動
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("サーバー起動エラー:%v", err)
	}
}

// --------------------------------------------------------------------------------------
// 関数名：	ConnectDB
// 機能：	データベースへ接続し、ハンドラーを返す
// 引数：	なし
// 戻り値：	データベースハンドラー
// --------------------------------------------------------------------------------------
func ConnectDB() *sql.DB {

	// 基準時刻を東京に設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalf("地域ロードエラー:%v", err)
	}

	// 事前にymlで指定したデータベース情報を使ってログイン
	// 本来なら外部ファイルや認証プログラムなどのある程度セキュリティに担保がある操作でログインしたいところ
	c := mysql.Config{
		DBName:    "db",
		User:      "user",
		Passwd:    "password",
		Addr:      "db:3306",
		Net:       "tcp",
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}
	db, err := sql.Open("mysql", c.FormatDSN())

	if err != nil {
		log.Fatalf("SQLの作成エラー:%v", err)
	}

	return db
}

// --------------------------------------------------------------------------------------
// 関数名：	CreateTable
// 機能：	レコードを保存するためのテーブルを作成
//			(未作成時のみ。作成済みの場合は何もせず)
// 引数：	*sql.DB		データベースハンドラー
// 戻り値：	なし
// --------------------------------------------------------------------------------------
func CreateTable(db *sql.DB) {

	// テーブル新規作成クエリを作成
	query := `CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT,
		name VARCHAR(255),
		age INT,
		date VARCHAR(255),
		PRIMARY KEY (id)
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("テーブル作成エラー: %v", err)
	}
}

// --------------------------------------------------------------------------------------
// 関数名：	RegisterUserData
// 機能：	フロントエンドから送られてきたユーザー情報をデータベースに登録
// 引数：	http.ResponseWriter		フロントエンドへ返すレスポンス
//			*http.Request			フロントエンドから受け取ったリクエスト情報
//			*sql.DB					データベースハンドラー
// 戻り値：	無し
// --------------------------------------------------------------------------------------
func RegisterUserData(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	// CORSヘッダーを設定
	SetHeader_CORS(w)

	// 本来はGETやPOSTの判別を行って、各リクエストに対応した処理を行う
	// 今回は仕様が決まっているので、POSTが来ること前提で進める
	if r.Method != http.MethodPost {
		http.Error(w, "今回はテストなのでPOSTメソッドのみ受け付けます", http.StatusMethodNotAllowed)
		return
	}

	// フォームデータを正常に受信できているかチェック
	// Goで受け取る時はParseFormではなくParseMultipartFormで受け取る
	//https://blog.84b9cb.info/posts/go-receive-form-data/
	err := r.ParseMultipartForm(1024 * 5) // サイズは適当なところで
	if err != nil {
		http.Error(w, "フォームデータの解析に失敗しました", http.StatusInternalServerError)
		return
	}

	// フォームからデータを取得
	name := r.FormValue("registerName")	// 名前
	age := r.FormValue("registerAge")	// 年齢

	// 年齢をint型に変換
	ageInt, err := strconv.Atoi(age)
	// フロントエンド側で入力typeをnumberにセットしているので、ないとは思うがここでも一応キャストが出来たかチェック
	if err != nil {
		log.Fatalf("年齢のキャストエラー:%v", err)
	}

	// 現在の日時を取得
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// データベースに登録
	user := User{Name: name, Age: ageInt, Date: currentTime}
	RegisterUser(db, &user)

	// ※デバッグ用
	// 登録したデータを読み出してコンソールに出力
	//id := RegisterUser(db, &user)
	//user.ID = id
	//ReadUser(db, user.ID)

	// 成功レスポンスを返して終了
	// (レスポンスでデータを返すので、リダイレクトはjs側で行う)
	w.WriteHeader(http.StatusOK)
}


// --------------------------------------------------------------------------------------
// 関数名：	ReadUserData
// 機能：	データベースに登録されたユーザー情報を全て取得し、JSON形式にしてフロントエンドへ返す
// 引数：	http.ResponseWriter		フロントエンドへ返すレスポンス
//			*http.Request			フロントエンドから受け取ったリクエスト情報
//			*sql.DB					データベースハンドラー
// 戻り値：	無し
// --------------------------------------------------------------------------------------
// フロントエンドからデータベース読み込み要求に対応
func ReadUserData(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	// CORSヘッダーを設定
	SetHeader_CORS(w)

	// データベースから全ユーザー情報をスライスで取得
	users := GetAllUsers(db)

	// 1件も登録が無かったらこのタイミングで例として1件登録しておく
	if len(users) == 0 {
		AddDefaultUserData(db)
		// 登録したら再度スライスで取得
		users = GetAllUsers(db)
	}

	// スライスで受け取ったものをjsonへ
	jsonResponse, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 作成したjsonをレスポンスで返す
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

	// 成功レスポンスを返して終了
	w.WriteHeader(http.StatusOK)
}

// --------------------------------------------------------------------------------------
// 関数名：	DeleteUserData
// 機能：	フロントエンドから指定のあったIDのレコードを削除
// 引数：	http.ResponseWriter		フロントエンドへ返すレスポンス
//			*http.Request			フロントエンドから受け取ったリクエスト情報
//			*sql.DB					データベースハンドラー
// 戻り値：	無し
// --------------------------------------------------------------------------------------
func DeleteUserData(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	// CORSヘッダーを設定
	SetHeader_CORS(w)

	// フロントエンドから削除を行いたいレコードのIDを取得
	unique_ID := r.FormValue("ID")

	// int64型に変換
	unique_id, err := strconv.ParseInt(unique_ID, 10, 64)
	if err != nil {
		// エラーハンドリング
		log.Fatalf("フロントエンドから取得した一意のIDの取得エラー:%v", err)
	}

	// 受け取ったIDを持つレコードの削除処理を実行
	DeleteUser(db, unique_id)

	// 1件も登録が無かったらこのタイミングで例として1件登録しておく
	users := GetAllUsers(db)
	if len(users) == 0 {
		// デフォのユーザー情報を登録
		AddDefaultUserData(db)

		// alertで最低1件の追加が必要なのを通知したいので異なるレスポンスを返す
		w.WriteHeader(http.StatusCreated) // 他の数字も試したけど、フロントエンドで受け取れる数字が限られる？とりあえず201を返すけど、本来の意味と違うかも…。
	} else {
		// 成功レスポンスを返して終了
		w.WriteHeader(http.StatusOK)
	}
}

// --------------------------------------------------------------------------------------
// 関数名：	SetHeader_CORS
// 機能：	CORSヘッダーを設定
// 引数：	http.ResponseWriter		フロントエンドへ返すレスポンス
// 戻り値：	無し
// --------------------------------------------------------------------------------------
func SetHeader_CORS(w http.ResponseWriter) {
	// CORSヘッダーを設定
	// https://developer.mozilla.org/ja/docs/Web/HTTP/CORS
	// CORSポリシーによって異なるオリジン間の通信がブロックされる。今回はテストなのですべてのオリジンからのアクセスを許可する
	w.Header().Set("Access-Control-Allow-Origin", "*")                   // すべてのオリジンからのアクセスを許可
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") // 許可するHTTPメソッド
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // 許可するヘッダー
}

// --------------------------------------------------------------------------------------
// 関数名：	GetAllUsers
// 機能：	データベースに登録された全レコード情報を取得してスライスで返す
// 引数：	*sql.DB		データベースハンドラー
// 戻り値：	全レコード情報を格納したスライス
// --------------------------------------------------------------------------------------
func GetAllUsers(db *sql.DB) []User {

	// このスライスにユーザー情報を入れていく
	users := []User{}

	// データベースから全ユーザー情報を取得するクエリを実行
	query := `SELECT id, name, age, date FROM users`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("全ユーザー情報を取得するクエリの失敗: %v", err)
		return nil
	}
	defer rows.Close()

	// 取得した行を一つずつ読み出してスライスを作成
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Date); err != nil {
			log.Fatalf("ユーザー情報のスキャンエラー: %v", err)
			return nil
		}
		users = append(users, user)
	}

	return users
}

// --------------------------------------------------------------------------------------
// 関数名：	RegisterUser
// 機能：	user情報を受け取って、データベースにレコード追加
// 引数：	*sql.DB		データベースハンドラー
//			*User		ユーザー情報
// 戻り値：	レコードの一意のID
// --------------------------------------------------------------------------------------
func RegisterUser(db *sql.DB, user *User) int64 {

	// ユーザー情報をレコード追加するためのクエリを作成
	query := `INSERT INTO users (name, age, date) VALUES (?, ?, ?)`
	result, err := db.Exec(query, user.Name, user.Age, user.Date)
	if err != nil {
		log.Fatalf("ユーザー登録エラー: %v", err)
		return 0
	}

	// 今登録したデータの一意のIDを取得
	// このIDをフロントエンド側のテーブルに覚えておいて後から一対一で対応できるようにする
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("最後に登録した情報のID取得エラー: %v", err)
		return 0
	}

	return id
}

// --------------------------------------------------------------------------------------
// 関数名：	ReadUser
// 機能：	ユーザーから指定のあったIDのレコードを読みだしてコンソールに出力	※デバッグ用
// 引数：	*sql.DB		データベースハンドラー
//			id			読み出したいレコードの一意のID
// 戻り値：	なし
// --------------------------------------------------------------------------------------
func ReadUser(db *sql.DB, id int64) {

	// 読みだしたレコードの格納箱
	var user User

	// 読み出しクエリの作成
	query := "SELECT id, name, age, date FROM users WHERE id = ?"
	// 作成したクエリでレコード読み出し
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Age, &user.Date)
	if err != nil {
		log.Fatalf("ユーザー読み出しエラー: %v", err)
	}
	// 読みだしたレコードをコンソールに出力
	fmt.Printf("読み出したユーザー: %#v\n", user)
}

// --------------------------------------------------------------------------------------
// 関数名：	DeleteUser
// 機能：	ユーザーから指定のあったIDのレコードを削除
// 引数：	*sql.DB		データベースハンドラー
//			id			削除したいレコードの一意のID
// 戻り値：	なし
// --------------------------------------------------------------------------------------
func DeleteUser(db *sql.DB, id int64) {
	// GolangのQueryとExecについて
	// https://sourjp.github.io/posts/go-db/

	// 削除クエリを作成
	query := "DELETE FROM users WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// 削除を実行
	_, err = stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}
}

// --------------------------------------------------------------------------------------
// 関数名：	AddDefaultUserData
// 機能：	登録件数が0件になってしまったときのデフォとしてのレコード1件追加処理
// 引数：	*sql.DB		データベースハンドラー
// 戻り値：	なし
// --------------------------------------------------------------------------------------
func AddDefaultUserData(db *sql.DB) {

	// 名前も年齢も決め打ちで
	name := "田中太郎"
	age := 20

	// 現在の日時を取得
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// データベースに登録
	user := User{Name: name, Age: age, Date: currentTime}
	RegisterUser(db, &user)
}