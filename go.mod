module github.com/refaktor/rye

go 1.26.1

retract v0.0.11 // Published accidentally with a bug

// replace github.com/refaktor/keyboard => ../keyboard

require (
	filippo.io/age v1.3.1
	github.com/BrianLeishman/go-imap v0.1.28
	github.com/JohannesKaufmann/html-to-markdown/v2 v2.5.1
	github.com/PuerkitoBio/goquery v1.12.0
	github.com/atotto/clipboard v0.1.4
	github.com/aws/aws-sdk-go-v2 v1.41.9
	github.com/aws/aws-sdk-go-v2/config v1.32.20
	github.com/aws/aws-sdk-go-v2/service/s3 v1.102.2
	github.com/aws/aws-sdk-go-v2/service/ses v1.34.26
	github.com/bitfield/script v0.24.1
	github.com/blevesearch/bleve/v2 v2.6.0
	github.com/blevesearch/bleve_index_api v1.3.12
	github.com/dgraph-io/badger/v4 v4.9.1
	// NOTE: Vulnerabilities GHSA-x744-4wpc-v9h2 and GHSA-pjbgf-vjq9-x5w7 are server-side only, client usage has limited risk
	github.com/drewlanenga/govector v0.0.0-20220726163947-b958ac08bc93
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/fsnotify/fsnotify v1.10.1
	github.com/fxtlabs/primes v0.0.0-20150821004651-dad82d10a449
	github.com/gdamore/tcell/v2 v2.13.10
	github.com/gen2brain/go-mpv v0.2.3
	github.com/glebarez/sqlite v1.11.0
	github.com/gliderlabs/ssh v0.3.8
	github.com/go-echarts/go-echarts/v2 v2.7.2
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-sql-driver/mysql v1.10.0
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gobwas/ws v1.4.0
	github.com/gorilla/sessions v1.4.0
	github.com/hpcloud/tail v1.0.0
	github.com/jinzhu/copier v0.4.0
	github.com/jlaffaye/ftp v0.2.1-0.20251026020404-6602e981a1bb
	github.com/jwalton/go-supportscolor v1.2.0
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54
	github.com/labstack/echo v3.3.10+incompatible
	github.com/landlock-lsm/go-landlock v0.8.1
	github.com/lib/pq v1.12.3
	github.com/mattn/go-runewidth v0.0.24
	github.com/mhale/smtpd v0.8.3
	github.com/mrz1836/postmark v1.9.2
	github.com/muesli/reflow v0.3.0
	github.com/ollama/ollama v0.30.0
	github.com/openai/openai-go v1.12.0
	github.com/pkg/term v1.2.0-beta.2.0.20211217091447-1a4a3b719465
	github.com/prometheus/client_golang v1.23.2
	github.com/refaktor/go-find v0.0.1
	github.com/refaktor/keyboard v0.0.0-20260517095250-755a59d30156
	github.com/sairash/chitosocket v1.0.2
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/thomasberger/parsemail v1.2.7
	github.com/xuri/excelize/v2 v2.10.1
	github.com/yuin/goldmark v1.8.2
	go.mongodb.org/mongo-driver v1.17.9
	golang.org/x/crypto v0.52.0
	golang.org/x/net v0.55.0
	golang.org/x/sync v0.20.0
	golang.org/x/term v0.43.0
	golang.org/x/text v0.37.0
	gopkg.in/headzoo/surf.v1 v1.0.1
	gopkg.in/yaml.v3 v3.0.1
	software.sslmate.com/src/go-pkcs12 v0.7.1
)

require (
	github.com/mlange-42/ark v0.8.3
	github.com/spf13/cobra v1.10.2
	github.com/tliron/glsp v0.2.2
	google.golang.org/genai v1.59.0
)

require (
	bazil.org/fuse v0.0.0-20230120002735-62a210ff1fd5 // indirect
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	filippo.io/hpke v0.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.4 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azfile v1.5.4 // indirect
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/FilenCloudDienste/filen-sdk-go v0.0.39 // indirect
	github.com/Files-com/files-sdk-go/v3 v3.3.82 // indirect
	github.com/GianlucaP106/gotmux v0.5.0 // indirect
	github.com/IBM/go-sdk-core/v5 v5.21.2 // indirect
	github.com/Max-Sum/base32768 v0.0.0-20230304063302-18e6ce5945fd // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/bcrypt v0.0.0-20211005172633-e235017c1baf // indirect
	github.com/ProtonMail/gluon v0.17.1-0.20230724134000-308be39be96e // indirect
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/ProtonMail/go-mime v0.0.0-20230322103455-7d82a3887f2f // indirect
	github.com/ProtonMail/go-srp v0.0.7 // indirect
	github.com/ProtonMail/gopenpgp/v2 v2.9.0 // indirect
	github.com/RoaringBitmap/roaring/v2 v2.14.5 // indirect
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/StirlingMarketingGroup/go-retry v0.0.0-20190512160921-94a8eb23e893 // indirect
	github.com/a1ex3/zstd-seekable-format-go/pkg v0.10.0 // indirect
	github.com/a8m/tree v0.0.0-20240104212747-2c8764a5f17e // indirect
	github.com/aalpar/deheap v1.1.2 // indirect
	github.com/abbot/go-http-auth v0.4.0 // indirect
	github.com/adrg/xdg v0.5.3 // indirect
	github.com/anacrolix/dms v1.7.2 // indirect
	github.com/anacrolix/generics v0.2.0 // indirect
	github.com/anacrolix/log v0.17.0 // indirect
	github.com/anchore/go-lzo v0.1.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/appscode/go-querystring v0.0.0-20170504095604-0126cfb3f1dc // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.22.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.1.1 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blevesearch/zapx/v17 v17.1.2 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/bradenaw/juniper v0.15.3 // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/buengese/sgzip v0.1.1 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/calebcase/tmpfile v1.0.3 // indirect
	github.com/cention-sany/utf7 v0.0.0-20170124080048-26cad61bd60a // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chilts/sid v0.0.0-20190607042430-660e94789ec9 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/cloudinary/cloudinary-go/v2 v2.15.0 // indirect
	github.com/cloudsoda/go-smb2 v0.0.0-20250228001242-d4c70e6251cc // indirect
	github.com/cloudsoda/sddl v0.0.0-20250224235906-926454e91efc // indirect
	github.com/colinmarc/hdfs/v2 v2.4.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.6.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/creasty/defaults v1.8.0 // indirect
	github.com/cronokirby/saferith v0.33.1-0.20250226174546-1f11f94ce488 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/diskfs/go-diskfs v1.7.0 // indirect
	github.com/dromara/dongle v1.0.1 // indirect
	github.com/dropbox/dropbox-sdk-go-unofficial/v6 v6.0.5 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/emersion/go-message v0.18.2 // indirect
	github.com/emersion/go-vcard v0.0.0-20241024213814-c9703dde27ff // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/flynn/noise v1.1.0 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/geoffgarside/ber v1.2.0 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/go-darwin/apfs v0.0.0-20211011131704-f84b94dbf348 // indirect
	github.com/go-git/go-billy/v5 v5.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/errors v0.22.6 // indirect
	github.com/go-openapi/strfmt v0.25.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.14 // indirect
	github.com/googleapis/gax-go/v2 v2.21.0 // indirect
	github.com/gorilla/schema v1.4.1 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hanwen/go-fuse/v2 v2.9.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/headzoo/surf v1.0.1 // indirect
	github.com/headzoo/ut v0.0.0-20181013193318-a13b5a7a02ca // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inbucket/html2text v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/internxt/rclone-adapter v0.0.0-20260331173834-036f908d0160 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/goidentity/v6 v6.0.1 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jhillyerd/enmime/v2 v2.3.0 // indirect
	github.com/jtolio/noiseconn v0.0.0-20231127013910-f6d9ecbf1de7 // indirect
	github.com/jzelinskie/whirlpool v0.0.0-20201016144138-0675e54bb004 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/koofr/go-httpclient v0.0.0-20240520111329-e20f8f203988 // indirect
	github.com/koofr/go-koofrclient v0.0.0-20221207135200-cbd7fc9ad6a6 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lanrat/extsort v1.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lpar/date v1.0.0 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mholt/archives v0.1.5 // indirect
	github.com/mikelolasagasti/xz v1.0.1 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/minio/xxml v0.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncw/swift/v2 v2.0.5 // indirect
	github.com/nwaples/rardecode/v2 v2.2.2 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/errors v1.2.0 // indirect
	github.com/olekukonko/ll v0.1.8 // indirect
	github.com/olekukonko/tablewriter v1.1.4 // indirect
	github.com/oracle/oci-go-sdk/v65 v65.111.0 // indirect
	github.com/panjf2000/ants/v2 v2.11.5 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pengsrc/go-shared v0.2.1-0.20190131101655-1999055a4a14 // indirect
	github.com/peterh/liner v1.2.2 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.10 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/otp v1.5.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/putdotio/go-putio/putio v0.0.0-20200123120452-16d982cac2b8 // indirect
	github.com/rasky/go-xdr v0.0.0-20170124162913-1a41d1a06c93 // indirect
	github.com/rclone/Proton-API-Bridge v1.0.3 // indirect
	github.com/rclone/go-proton-api v1.0.2 // indirect
	github.com/rclone/gofakes3 v0.0.6 // indirect
	github.com/rclone/rclone v1.74.3 // indirect
	github.com/relvacode/iso8601 v1.7.0 // indirect
	github.com/rfjakob/eme v1.2.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	github.com/shabbyrobe/gocovmerge v0.0.0-20230507112040-c3350d9342df // indirect
	github.com/shirou/gopsutil/v4 v4.26.3 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/sourcegraph/jsonrpc2 v0.2.0 // indirect
	github.com/spacemonkeygo/monkit/v3 v3.0.25-0.20251022131615-eb24eb109368 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/sqs/go-xoauth2 v0.0.0-20120917012134-0911dad68e56 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/t3rm1n4l/go-mega v0.0.0-20251120131202-6845944c051c // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tliron/commonlog v0.2.8 // indirect
	github.com/tliron/kutil v0.3.11 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/unknwon/goconfig v1.0.0 // indirect
	github.com/willscott/go-nfs v0.0.4 // indirect
	github.com/willscott/go-nfs-client v0.0.0-20251022144359-801f10d98886 // indirect
	github.com/winfsp/cgofuse v1.6.1-0.20260126094232-f2c4fccdb286 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yunify/qingstor-sdk-go/v3 v3.2.0 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org v0.0.0-20260112195520-a5071408f32f // indirect
	goftp.io/server/v2 v2.0.2 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/image v0.41.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/api v0.275.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
	google.golang.org/grpc v1.81.1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.77 // indirect
	moul.io/http2curl/v2 v2.3.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
	storj.io/common v0.0.0-20260225132117-99155641c30a // indirect
	storj.io/drpc v0.0.35-0.20250513201419-f7819ea69b55 // indirect
	storj.io/eventkit v0.0.0-20250410172343-61f26d3de156 // indirect
	storj.io/infectious v0.0.2 // indirect
	storj.io/picobuf v0.0.4 // indirect
	storj.io/uplink v1.14.0 // indirect
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/JohannesKaufmann/dom v0.2.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.11 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.19 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.36.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.3 // indirect
	github.com/aws/smithy-go v1.26.0 // indirect
	github.com/bits-and-blooms/bitset v1.24.2 // indirect
	github.com/blevesearch/geo v0.2.5 // indirect
	github.com/blevesearch/go-faiss v1.1.0 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/gtreap v0.1.1 // indirect
	github.com/blevesearch/mmap-go v1.2.0 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.4.7 // indirect
	github.com/blevesearch/segment v0.9.1 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.2 // indirect
	github.com/blevesearch/vellum v1.2.0 // indirect
	github.com/blevesearch/zapx/v11 v11.4.3 // indirect
	github.com/blevesearch/zapx/v12 v12.4.3 // indirect
	github.com/blevesearch/zapx/v13 v13.4.3 // indirect
	github.com/blevesearch/zapx/v14 v14.4.3 // indirect
	github.com/blevesearch/zapx/v15 v15.4.3 // indirect
	github.com/blevesearch/zapx/v16 v16.3.4 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elastic/go-seccomp-bpf v1.6.0
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/itchyny/gojq v0.12.13 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/labstack/gommon v0.4.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260216142805-b3301c5f2a88 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/sys v0.45.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gorm.io/gorm v1.25.7 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
	mvdan.cc/sh/v3 v3.7.0 // indirect
)
