package functions

import (
        "context"
        "fmt"
        "log"
        "net/http"
        "os"
        "time"

        //Firestoreの操作に必要なライブ
        "cloud.google.com/go/firestore"
)

//Firestoreにデータを追加するための構造体、タグで変数とキーを紐づける
type Word struct {
        Word        string    `firestore:"WORD"`
        Description string    `firestore:"DESCRIPTION"`
        Datetime    time.Time `firestore:"DATETIME"`
}

//HTTPトリガーで実行される
// Wtabb -> What's this abbreviation?
// 略語: abbreviation
func Wtabb(w http.ResponseWriter, r *http.Request) {
        //コンテキストを取得する
        ctx := context.Background()
        //プロジェクトIDを取得する
        projectID := os.Getenv("GCP_PROJECT_ID")

        switch r.Method {
        case http.MethodPost: //POSTの場合

                //パラメータから値を取り出す
                word := r.PostFormValue("word")
                desc := r.PostFormValue("description")
                abb := r.PostFormValue("abb")

                //取り出せない場合はエラーとして処理を終了する
                if word == "" {
                        fmt.Fprint(w, "パラメータに\"word\"がありません。\r\n")
                        return
                }
                if desc == "" {
                        fmt.Fprint(w, "パラメータに\"desc\"がありません。\r\n")
                        return
                }
                if abb == "" {
                        abb = word
                }

                struct_word := Word{}
                struct_word.Word = word
                struct_word.Description = desc
                //Firestoreへ出力する関数を呼び出す
                CreateWordFirestore(struct_word, abb, ctx, projectID)
                w.Write([]byte("Registered Word!"))

        case http.MethodGet: //Getの場合
                abb := r.URL.Query().Get("abb")
                if abb == "" {
                        fmt.Fprint(w, "パラメータに\"abb\"がありません。\r\n")
                        return
                }
                word := Word{}
                _, word = GetWordFirestore(abb, ctx, projectID)
                if word.Word == "" && word.Description == "" {
                        fmt.Fprint(w, "wordとdescriptionが見つかりませんでした :(")
                } else {
                        fmt.Fprint(w, word.Word)
                        fmt.Fprint(w, " / ")
                        fmt.Fprint(w, word.Description)
                }
                w.Write([]byte(" ;"))

        default: //GET,POST以外の場合はエラー
                http.Error(w, "405 - Method Not Allowed", http.StatusMethodNotAllowed)
        }
}

func GetWordFirestore(key string, ctx context.Context, projectID string) (error, Word) {
        ent := Word{}

        //Firestoreを操作するクライアントを作成する、エラーの場合はLoggingへ出力する
        client, err := firestore.NewClient(ctx, projectID)
        if err != nil {
                log.Printf("Firestore接続エラー　Error:%T message: %v", err, err)
                return err, ent
        }
        defer client.Close()

        dsnap, err := client.Collection("WORDS").Doc(key).Get(ctx)
        if err != nil {
            return err, ent
        }
        if err = dsnap.DataTo(&ent); err != nil {
          log.Printf("マッピングエラー Error:%T", err)
        }
        return nil, ent
}

func CreateWordFirestore(word Word, key string, ctx context.Context, projectID string) {

        //Firestoreを操作するクライアントを作成する、エラーの場合はLoggingへ出力する
        client, err := firestore.NewClient(ctx, projectID)
        if err != nil {
                log.Printf("Firestore接続エラー　Error:%T message: %v", err, err)
                return
        }

        //確実にクライアントを閉じるようにする
        defer client.Close()

        //現在時刻を構造体へ格納する
        word.Datetime = time.Now()

        //Firestoreの追加を行う、エラーの場合はLoggingへ出力する
        _, err = client.Collection("WORDS").Doc(key).Set(ctx, word)
        if err != nil {
                log.Printf("データ書き込みエラー　Error:%T message: %v", err, err)
                return
        }
}
