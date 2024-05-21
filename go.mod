module github.com/refaktor/rye

go 1.21

// toolchain go1.21.5

retract v0.0.11 // Published accidentally with a bug

require (
	github.com/aws/aws-sdk-go-v2 v1.26.1
	github.com/aws/aws-sdk-go-v2/config v1.27.11
	github.com/aws/aws-sdk-go-v2/service/ses v1.22.4
	github.com/bitfield/script v0.22.0
	github.com/blevesearch/bleve/v2 v2.4.0
	github.com/blevesearch/bleve_index_api v1.1.6
	github.com/drewlanenga/govector v0.0.0-20220726163947-b958ac08bc93
	github.com/gliderlabs/ssh v0.3.7
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-sql-driver/mysql v1.8.1
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gobwas/ws v1.3.2
	github.com/gorilla/sessions v1.2.2
	github.com/jinzhu/copier v0.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/lib/pq v1.10.9
	github.com/mattn/go-runewidth v0.0.15
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/mhale/smtpd v0.8.2
	github.com/mrz1836/postmark v1.6.4
	github.com/pkg/term v1.1.0
	github.com/refaktor/go-peg v0.0.0-20220116201714-31e3dfa8dc7d
	github.com/refaktor/liner v1.2.10
	github.com/sashabaranov/go-openai v1.20.4
	github.com/shirou/gopsutil/v3 v3.24.4
	github.com/thomasberger/parsemail v1.2.6
	go.mongodb.org/mongo-driver v1.14.0
	golang.org/x/crypto v0.23.0
	golang.org/x/net v0.24.0
	golang.org/x/sync v0.7.0
	golang.org/x/term v0.20.0
	golang.org/x/text v0.15.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/RoaringBitmap/roaring v1.2.3 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.6 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/blevesearch/geo v0.1.20 // indirect
	github.com/blevesearch/go-faiss v1.0.13 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.0.4 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.2.9 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.0.10 // indirect
	github.com/blevesearch/zapx/v11 v11.3.10 // indirect
	github.com/blevesearch/zapx/v12 v12.3.10 // indirect
	github.com/blevesearch/zapx/v13 v13.3.10 // indirect
	github.com/blevesearch/zapx/v14 v14.3.10 // indirect
	github.com/blevesearch/zapx/v15 v15.3.13 // indirect
	github.com/blevesearch/zapx/v16 v16.0.12 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/itchyny/gojq v0.12.12 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/labstack/gommon v0.4.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/yhirose/go-peg v0.0.0-20210804202551-de25d6753cf1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	golang.org/x/sys v0.20.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	mvdan.cc/sh/v3 v3.6.0 // indirect
)
