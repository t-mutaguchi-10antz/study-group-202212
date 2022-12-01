---
marp: true
theme: gaia
_class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')

---
# go ctx & err

mutaguchi

![bg right:40% 80%](https://camo.qiitausercontent.com/3d8be28fc36d7110e84703b928077a3ec66292c7/68747470733a2f2f71696974612d696d6167652d73746f72652e73332e616d617a6f6e6177732e636f6d2f302f31343935322f35633835316335352d393863392d306235622d356532312d3339633133366130353834342e706e67)


---
## アジェンダ

- テーマの選定理由
- Context
- Error
- おまけ


---
## テーマの選定理由

(作図する)

- 聞いてくれている方にも
- 自身にとっても学びになる


---
## Context

ネットワーク I/O が発生する箇所では必ず使う

```go
ctx := context.Background()
client, err := spanner.NewClient(ctx, "projects/foo/instances/bar/databases/zoo")
```

---
## Go の並行処理パターン - Context
 
- Go 製サーバーは各リクエストを独自ゴルーチンで処理する
- 一連のゴルーチンはリクエスト固有値にアクセスする必要ある
	- エンドユーザ ID／承認トークン／リクエスト期限など
- リクエストがキャンセル ( デッドラインやタイムアウトを含む ) した場合、動作しているすべてのゴルーチンはすぐに終了し、システムが使用中のリソースを再利用できるようにする必要がある

これらの解決手段として context パッケージが用意されている

cf. [Go Concurrency Patterns: Context](https://go.dev/blog/context)


---
## 責務 - Context

- キャンセルの伝達
- リクエスト固有値の伝達

cf. [Standard library > context - Overview](https://pkg.go.dev/context#pkg-overview)


---
## 責務 - Context

デッドラインもタイムアウトもキャンセルをラップした処理なので、本スライドでは一括してキャンセル処理として扱う

```go
func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
	( ... 略 ... )
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// The current deadline is already sooner than the new one.
		return WithCancel(parent)
	}
	( ... 略 ... )
}
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}
```

---
## 責務ではない事 - Context

- 関数のオプショナル引数 ( ≒ python のキーワード引数 ) ではない
- WithValue をオプショナル引数用途で使うと関数の実行に必要なシグネチャが分からなくなる

NG )

```golang
func rename(ctx context.Context) error {
	name := ctx.Value("name")
}
```

cf. [Standard library > context - Overview](https://pkg.go.dev/context#pkg-overview)


---
## 責務ではない事 - Context

- オプショナル引数が必要なら Functional Option Pattern ( FOP ) の採用を検討する

```go
options := []option.ClientOption{
	option.WithCredentialsFile("PATH_TO_CREDENTIALS_FILE"),
}
client, err := spanner.NewClient(ctx, dbName, options...)
```

FOP も乱用するとシグネチャが分からなくなるので、必須パラメータを FOP にしたりしない事

---
## データ構造 - Context

```go
type valueCtx struct {
	Context
}
```

- 隣接リストで実装されている
- 値の参照は、子から親のコンテキストを参照可能
- キャンセル処理の伝達は、親から子へのみ行われる
	- 親がキャンセルされた → 子もキャンセルされる
	- 子がキャンセルされた → 親はキャンセルされない

cf. [Goのcontext.Contextで学ぶ有向グラフと実装](https://future-architect.github.io/articles/20210629a/)

---
## インターフェース - Context

```go
type Context interface {
	// Done は、この Context がキャンセルされるか、タイムアウトしたときに閉じられるチャネルを返す
	Done() <-chan struct{}
	// Err は、Done チャンネルが閉じた後、このコンテキストがキャンセルされた理由を示す
	Err() error
	// Deadline は、この Context がキャンセルされる時刻（がもしあれば）を返す
	Deadline() (deadline time.Time, ok bool)
	// Value は、key に関連する値を返し、無い場合は nil を返す
	Value(key any) any
}
```

キャンセルされるとチャンネルを使った伝達が行われる


---
## 動作確認 - Context

- ケース 1. 直列の場合

```
% go run cmd/ctx_cancel_1/main.go
canceled -> 3
canceled -> 2
```

- ケース 2. 木構造の場合

```
% go run cmd/ctx_cancel_2/main.go
canceled -> 1-2
canceled -> 1-1
```


---
## チャンネル - Context

なぜ Go では並列処理の信号伝達にチャンネルを使うのか？

紐解くと `C10K問題` という世界で最も普及した Apache サーバーが抱える問題に行き当たる


---
## C10K問題 - Context

- クライアントが 10,000 台を超えるとプロセス数上限に達する
  - Apache は 1 リクエスト 1 プロセス ( Apache 方式 )
	- 32bit Linux ではプロセス数上限が 32,767 であるため、それ以上リクエストを捌けなくなる

- コンテキストスイッチのコストが増大
	- CPU が複数プロセスを並行処理するため、それまでの処理内容を保存して新しい処理の内容を復元すること
	- Apache 方式ではリクエスト増＝プロセス増であるため、コンテキストスイッチのコストが無視できなくなる


---
## C10K問題 - Context

これらの問題はシングルプロセス・マルチスレッドにすればかなり軽減されるらしいが、それでもファイルディスクリプタ上限の問題が残る

cf. [【勉強メモ】C10K問題【マルチプロセス・マルチスレッド】](https://udon-yuya.hatenablog.com/entry/2020/09/03/233227)


---
## 並行プログラミング - Context

C10K問題に代表される有限なプロセス・スレッドの活用問題に対し、モダンなプログラミング言語ではスレッドを効率的に利用する機構が用意されている

Go は開発者である Rob Pike 氏が Communicating Sequential Processes ( CSP ) モデルが基になっていると述べており、ゴルーチンとチャンネルによって実現されている

cf. [Origins of Go Concurrency style by Rob Pike](https://youtu.be/3DtUzH3zoFo?t=130)


---
## channel or async/await - Context

並行プログラミングの処理系としては様々な言語で async/await が採用されているが、Go の開発者たちはデメリットも多いと考え CSP に基づく方法が採用された

( 公式な情報元へは辿り着けなかったので各自でご判断ください )

cf. [Goはなぜasync/awaitを採用しなかったの？](https://zenn.dev/nobonobo/articles/9a9f12b27bfde9#go%E3%81%AF%E3%81%AA%E3%81%9Casync%2Fawait%E3%82%92%E6%8E%A1%E7%94%A8%E3%81%97%E3%81%AA%E3%81%8B%E3%81%A3%E3%81%9F%E3%81%AE%EF%BC%9F)


---
## まとめ - Context

- キャンセルの伝達
  - 並行処理を行う上でスレッドを有効活用し、かつ、シンプルに表現できる CSP スタイルが基となり、ゴルーチン＆チャンネルによって実現されているœ
- リクエスト固有値の伝達
	- グローバルな値 ( DB コネクションなど ) の受け渡しには使わない

#### 責務を理解して適切に取り扱おう

