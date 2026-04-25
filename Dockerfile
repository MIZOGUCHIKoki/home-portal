# ベースイメージの指定（バージョンを固定することで再現性を担保）
FROM golang:1.22-alpine

# 環境変数の設定
# CGOを無効化（Alpineはmusl libcを使用しているため、CGOを有効にするとビルドが複雑になることがある）
ENV TZ=Asia/Tokyo
ENV ROOT=/app
ENV CGO_ENABLED=0 


# 必要なパッケージのインストール
# tzdata: タイムゾーン設定用
# git: Goモジュールのダウンロード等で必要になる場合があるため
RUN apk update && apk add --no-cache \
	git \
	tzdata

# 作業ディレクトリの設定
WORKDIR ${ROOT}

# ※注意: 開発用なので COPY . . や go build は書きません。
# コードは compose.yaml の volumes でマウント（リンク）させます。