# drawio-to-html

編集用データを含む draw.io SVG を、Notion の HTML ブロックへアップロードできる HTML に変換する Go CLI です。

```sh
drawio-to-html ./sample.drawio.svg
# ./sample.drawio.svg.html を作成
```

元の SVG を画像として埋め込み、SVG の `content` 属性から抽出した draw.io XML を標準ビューアーへ渡します。ページ・レイヤー・スタイル・圧縮データ・ページ ID を保持し、下部ツールバーからズーム・フィット・移動・ページ・レイヤーを操作します。

## ビルド・インストール

Go **1.27.4** を使用します。

`go install` でインストールできます。

```sh
go install github.com/yuyj109/drawio-to-html@latest
drawio-to-html ./sample.drawio.svg
```

インストール先（`GOBIN`、未設定の場合は `$(go env GOPATH)/bin`）を `PATH` に追加してください。

ソースからビルド・インストールする場合：

```sh
make build
./bin/drawio-to-html ./sample.drawio.svg

# PATH 上の Go bin ディレクトリへインストールする場合
make install
drawio-to-html ./sample.drawio.svg
```

出力先は入力パスに `.html` を追加した場所です。成功時はそのパスを標準出力へ表示します。入力 SVG は変更しません。出力が既にある場合は `Overwrite? [y/N]` と確認し、`y` / `yes`（大文字可）を入力した場合だけ上書きします。Enter のみ・`n`・未入力での入力終了では変更せず、終了コード0でキャンセルします。確認は標準エラー出力へ表示します。`--force` は確認を省略します。

```sh
drawio-to-html --force ./sample.drawio.svg
drawio-to-html --help
```

HTML テンプレートはバイナリへ `go:embed` で同梱するため、実行時にテンプレートファイルや Node.js は不要です。CLI の変換中にネットワークへ接続することもありません。

## Notion で使う

1. draw.io の SVG エクスポートで **Include a copy of my diagram** を有効にします。
2. この CLI で SVG を HTML へ変換します。
3. Notion で `/html` を選び、生成された `.svg.html` をアップロードします。

編集用 XML がない通常の SVG、不正 XML、内部 DTD 定義、20 MiB を超える入力はエラーにします。SVG の標準的な DOCTYPE 宣言（PUBLIC / SYSTEM を含む）は許可し、外部 DTD・実体の取得や展開は行いません。圧縮・非圧縮のページ、URI エンコードされた `content` 属性に対応します。

生成 HTML は固定版 draw.io **v32.6.10** を jsDelivr CDN から取得します。ビューアーと図形定義バンドルを SRI（SHA-384）で検証し、AWS などのアイコンを描画します。draw.io 標準の相対画像パス（`img/lib/...`）も同じ固定版 CDN から読み込みます。CDN 読込みが遮断された場合や JavaScript が無効な場合は、HTML 内の元 SVG を画像として表示します。全画面表示が禁止される場合は Notion のブロックメニューを利用してください。実際の Notion 上での動作は未検証です。

生成 HTML に図のデータが含まれます。共有権限は Notion 側で管理してください。

## 構成

`../notion-sync/` に合わせ、Cobra のコマンドと変換処理を分離しています。

```text
main.go                              cmd.Execute の呼出し
cmd/root.go                          引数・フラグ・標準出力
internal/drawio/html.go               Render(svg, filename)：HTML を返す
internal/drawio/svg.go                SVG / 編集用 XML の検証・抽出
internal/drawio/file.go               ConvertFile(input, force)：ファイル変換
internal/drawio/templates/viewer.html.tmpl
examples/                            元の表示サンプル
```

`Render` はファイル保存やネットワーク通信を行いません。将来 notion-sync へ統合する際は `internal/drawio` を移植し、HTML 生成と Notion アップロードを分けて扱えます。

## 検証

```sh
make fmt        # go fmt
make lint       # fmt-check + golangci-lint + govulncheck
make test       # go test -race ./...
make check      # lint + test + build
```

golangci-lint は v2.15.0、govulncheck は v1.9.0 に固定。`.golangci.yml` は notion-sync と同じ構成です。

テストでは元 SVG・XML の保持、圧縮ページ、HTML / JSON のエスケープ、不正入力、上書き確認・キャンセル・確認の省略、同時作成を確認します。

draw.io の出典：[jgraph/drawio](https://github.com/jgraph/drawio)。固定ビューアーの URL と SRI は HTML テンプレート内で管理します。
