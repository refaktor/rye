//go:build !no_imap
// +build !no_imap

package batteries

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
)

// imapURI is the parsed form of an imap:// or imaps:// URI.
//
//	imap://user:password@host:143/INBOX    STARTTLS (port 993 switches to implicit TLS)
//	imaps://user:password@host:993/INBOX   implicit TLS
type imapURI struct {
	user        string
	password    string
	host        string
	port        int
	mailbox     string
	implicitTLS bool
}

func parseIMAPURI(scheme, path string) (*imapURI, error) {
	if scheme != "imap" && scheme != "imaps" {
		return nil, fmt.Errorf("URI scheme must be 'imap' or 'imaps'")
	}

	u, err := url.Parse(scheme + "://" + path)
	if err != nil {
		// url.Error includes the full URI, which may contain a password.
		if urlErr, ok := err.(*url.Error); ok {
			return nil, fmt.Errorf("invalid IMAP URI: %v", urlErr.Err)
		}
		return nil, fmt.Errorf("invalid IMAP URI")
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("IMAP URI must include a host")
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, fmt.Errorf("IMAP URI must not contain a query or fragment")
	}

	implicitTLS := scheme == "imaps"
	port := 0
	if p := u.Port(); p != "" {
		port, err = strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid IMAP port %q", p)
		}
	}
	if port == 0 {
		if implicitTLS {
			port = 993
		} else {
			port = 143
		}
	}
	// Port 993 is conventionally implicit TLS even when the imap:// scheme was
	// used. This keeps the familiar "Open imap://user:pass@host:993" form working.
	if !implicitTLS && port == 993 {
		implicitTLS = true
	}

	var user, password string
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}

	mailbox := strings.TrimPrefix(u.Path, "/")
	if mailbox != "" {
		mailbox, err = url.PathUnescape(mailbox)
		if err != nil {
			return nil, fmt.Errorf("invalid mailbox in IMAP URI: %w", err)
		}
	}

	return &imapURI{
		user:        user,
		password:    password,
		host:        u.Hostname(),
		port:        port,
		mailbox:     mailbox,
		implicitTLS: implicitTLS,
	}, nil
}

func imapOptions() *imapclient.Options {
	return &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	}
}

func dialIMAP(u *imapURI) (*imapclient.Client, error) {
	address := net.JoinHostPort(u.host, strconv.Itoa(u.port))
	if u.implicitTLS {
		return imapclient.DialTLS(address, imapOptions())
	}
	return imapclient.DialStartTLS(address, imapOptions())
}

func selectMailbox(client *imapclient.Client, mailbox string) error {
	if mailbox == "" {
		return nil
	}
	_, err := client.Select(mailbox, nil).Wait()
	return err
}

func openIMAP(u *imapURI, password, token, auth string) (*imapclient.Client, error) {
	if u.user == "" {
		return nil, fmt.Errorf("IMAP URI must include a username: imap://user[:pass]@host[:port]")
	}

	client, err := dialIMAP(u)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if auth == "oauth2" {
		err = authenticateIMAPOAuth2(client, u.user, token)
	} else {
		if password == "" {
			password = u.password
		}
		err = client.Login(u.user, password).Wait()
	}
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if err := selectMailbox(client, u.mailbox); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to select mailbox %q: %w", u.mailbox, err)
	}

	return client, nil
}

// xoauth2Client implements the XOAUTH2 SASL mechanism used by Gmail and several
// other IMAP providers. go-sasl only ships OAUTHBEARER, so provide the variant
// that is still most commonly advertised by mail servers.
type xoauth2Client struct {
	username string
	token    string
}

func (c *xoauth2Client) Start() (string, []byte, error) {
	return "XOAUTH2", []byte("user=" + c.username + "\x01auth=Bearer " + c.token + "\x01\x01"), nil
}

func (c *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	return nil, fmt.Errorf("XOAUTH2 authentication rejected: %s", string(challenge))
}

func authenticateIMAPOAuth2(client *imapclient.Client, username, token string) error {
	caps := client.Caps()
	if caps.Has(imap.AuthCap("XOAUTH2")) {
		return client.Authenticate(&xoauth2Client{username: username, token: token})
	}
	if caps.Has(imap.AuthCap("OAUTHBEARER")) {
		return client.Authenticate(sasl.NewOAuthBearerClient(&sasl.OAuthBearerOptions{
			Username: username,
			Token:    token,
		}))
	}
	// Some servers don't advertise the SASL mechanism until AUTHENTICATE is
	// attempted. Prefer XOAUTH2, which is the de-facto standard for IMAP OAuth2.
	return client.Authenticate(&xoauth2Client{username: username, token: token})
}

func imapClientFrom(arg env.Object) (*imapclient.Client, bool) {
	switch native := arg.(type) {
	case env.Native:
		client, ok := native.Value.(*imapclient.Client)
		return client, ok
	default:
		return nil, false
	}
}

func imapUIDsFromBlock(ps *env.ProgramState, block env.Block, fn string) ([]imap.UID, env.Object) {
	uids := make([]imap.UID, 0, block.Series.Len())
	for _, uidObj := range block.Series.S {
		uid, ok := uidObj.(env.Integer)
		if !ok {
			return nil, evaldo.MakeBuiltinError(ps, "All UIDs must be integers", fn)
		}
		if uid.Value <= 0 || uid.Value > int64(^uint32(0)) {
			return nil, evaldo.MakeBuiltinError(ps, fmt.Sprintf("Invalid UID %d", uid.Value), fn)
		}
		uids = append(uids, imap.UID(uint32(uid.Value)))
	}
	return uids, nil
}

func imapUIDSet(uids []imap.UID) imap.UIDSet {
	return imap.UIDSetNum(uids...)
}

func imapFlagsBlock(flags []imap.Flag) env.Object {
	objs := make([]env.Object, len(flags))
	for i, flag := range flags {
		objs[i] = *env.NewString(string(flag))
	}
	return *env.NewBlock(*env.NewTSeries(objs))
}

func imapAddressesString(addresses []imap.Address) string {
	parts := make([]string, 0, len(addresses))
	for _, address := range addresses {
		if address.IsGroupStart() || address.IsGroupEnd() {
			continue
		}
		addr := address.Addr()
		if addr == "" {
			continue
		}
		if address.Name != "" {
			parts = append(parts, fmt.Sprintf("%s <%s>", address.Name, addr))
		} else {
			parts = append(parts, addr)
		}
	}
	return strings.Join(parts, ", ")
}

func imapMailAddressesString(addresses []*mail.Address) string {
	parts := make([]string, 0, len(addresses))
	for _, address := range addresses {
		if address == nil {
			continue
		}
		parts = append(parts, address.String())
	}
	return strings.Join(parts, ", ")
}

func imapTimeString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func imapOverviewDict(msg *imapclient.FetchMessageBuffer) env.Dict {
	dict := env.NewDict(make(map[string]any))
	dict.Data["uid"] = *env.NewInteger(int64(msg.UID))
	dict.Data["size"] = *env.NewInteger(msg.RFC822Size)
	dict.Data["flags"] = imapFlagsBlock(msg.Flags)
	dict.Data["received"] = *env.NewString(imapTimeString(msg.InternalDate))

	var subject, from, to, cc, bcc, date, messageID string
	if envelope := msg.Envelope; envelope != nil {
		subject = envelope.Subject
		from = imapAddressesString(envelope.From)
		to = imapAddressesString(envelope.To)
		cc = imapAddressesString(envelope.Cc)
		bcc = imapAddressesString(envelope.Bcc)
		date = imapTimeString(envelope.Date)
		messageID = envelope.MessageID
	}

	dict.Data["subject"] = *env.NewString(subject)
	dict.Data["from"] = *env.NewString(from)
	dict.Data["to"] = *env.NewString(to)
	dict.Data["cc"] = *env.NewString(cc)
	dict.Data["bcc"] = *env.NewString(bcc)
	dict.Data["date"] = *env.NewString(date)
	dict.Data["message-id"] = *env.NewString(messageID)
	return *dict
}

func imapAttachmentFilename(header mail.PartHeader) string {
	if attachment, ok := header.(*mail.AttachmentHeader); ok {
		if filename, err := attachment.Filename(); err == nil && filename != "" {
			return filename
		}
	}
	if _, params, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil {
		if filename := params["filename"]; filename != "" {
			return filename
		}
	}
	if _, params, err := mime.ParseMediaType(header.Get("Content-Type")); err == nil {
		if filename := params["name"]; filename != "" {
			return filename
		}
	}
	return ""
}

func imapPartContentType(header mail.PartHeader) string {
	contentType, _, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil || contentType == "" {
		return header.Get("Content-Type")
	}
	return contentType
}

func imapAttachmentDict(header mail.PartHeader, body io.Reader) (env.Object, error) {
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	dict := env.NewDict(make(map[string]any))
	dict.Data["name"] = *env.NewString(imapAttachmentFilename(header))
	dict.Data["mime-type"] = *env.NewString(imapPartContentType(header))
	dict.Data["size"] = *env.NewInteger(int64(len(content)))
	// Rye has no binary type; keep the old behaviour and expose the raw content
	// as a string. Callers that need a file can write the string out directly.
	dict.Data["content"] = *env.NewString(string(content))
	return *dict, nil
}

func imapEmailDict(msg *imapclient.FetchMessageBuffer, raw []byte) env.Dict {
	dict := imapOverviewDict(msg)
	dict.Data["text"] = *env.NewString("")
	dict.Data["html"] = *env.NewString("")
	dict.Data["attachments"] = *env.NewBlock(*env.NewTSeries([]env.Object{}))

	if len(raw) == 0 {
		return dict
	}

	reader, _ := mail.CreateReader(bytes.NewReader(raw))
	if reader == nil {
		return dict
	}
	defer reader.Close()

	var text, html strings.Builder
	attachments := make([]env.Object, 0)

	if subject, err := reader.Header.Subject(); err == nil && subject != "" {
		dict.Data["subject"] = *env.NewString(subject)
	}
	if date, err := reader.Header.Date(); err == nil && !date.IsZero() {
		dict.Data["date"] = *env.NewString(imapTimeString(date))
	}
	if messageID, err := reader.Header.MessageID(); err == nil && messageID != "" {
		dict.Data["message-id"] = *env.NewString(messageID)
	}
	for key, field := range map[string]string{
		"from":     "From",
		"to":       "To",
		"cc":       "Cc",
		"bcc":      "Bcc",
		"sender":   "Sender",
		"reply-to": "Reply-To",
	} {
		if addresses, err := reader.Header.AddressList(field); err == nil && len(addresses) > 0 {
			dict.Data[key] = *env.NewString(imapMailAddressesString(addresses))
		}
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if part == nil {
			break
		}

		contentType := imapPartContentType(part.Header)
		switch part.Header.(type) {
		case *mail.InlineHeader:
			if strings.HasPrefix(contentType, "text/html") {
				if body, readErr := io.ReadAll(part.Body); readErr == nil {
					html.Write(body)
				}
			} else if strings.HasPrefix(contentType, "text/") || contentType == "" {
				if body, readErr := io.ReadAll(part.Body); readErr == nil {
					text.Write(body)
				}
			} else if attachment, attErr := imapAttachmentDict(part.Header, part.Body); attErr == nil {
				attachments = append(attachments, attachment)
			}
		case *mail.AttachmentHeader:
			if attachment, attErr := imapAttachmentDict(part.Header, part.Body); attErr == nil {
				attachments = append(attachments, attachment)
			}
		}

		if err != nil {
			break
		}
	}

	dict.Data["text"] = *env.NewString(text.String())
	dict.Data["html"] = *env.NewString(html.String())
	dict.Data["attachments"] = *env.NewBlock(*env.NewTSeries(attachments))
	return dict
}

func imapNewClient(ps *env.ProgramState, u *imapURI, password, token, auth, fn string) env.Object {
	client, err := openIMAP(u, password, token, auth)
	if err != nil {
		return evaldo.MakeBuiltinError(ps, err.Error(), fn)
	}
	return *env.NewNative(ps.Idx, client, "imap-client")
}

var imapSearchFlags = map[string]imap.Flag{
	"ANSWERED": imap.FlagAnswered,
	"DELETED":  imap.FlagDeleted,
	"DRAFT":    imap.FlagDraft,
	"FLAGGED":  imap.FlagFlagged,
	"RECENT":   imap.Flag("\\Recent"),
	"SEEN":     imap.FlagSeen,
}

func imapParseDate(value string) (time.Time, error) {
	for _, layout := range []string{"02-Jan-2006", "2-Jan-2006", "02-Jan-06", "2006-01-02"} {
		if date, err := time.Parse(layout, value); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid IMAP date %q", value)
}

func imapParseNumber(value string) (uint32, error) {
	if value == "*" {
		return 0, nil
	}
	number, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", value)
	}
	return uint32(number), nil
}

func imapParseNumSet(value string) ([][2]uint32, error) {
	var ranges [][2]uint32
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, fmt.Errorf("invalid empty set item in %q", value)
		}
		parts := strings.SplitN(item, ":", 2)
		start, err := imapParseNumber(parts[0])
		if err != nil {
			return nil, err
		}
		stop := start
		if len(parts) == 2 {
			stop, err = imapParseNumber(parts[1])
			if err != nil {
				return nil, err
			}
		}
		ranges = append(ranges, [2]uint32{start, stop})
	}
	return ranges, nil
}

func imapParseSeqSet(value string) (imap.SeqSet, error) {
	ranges, err := imapParseNumSet(value)
	if err != nil {
		return nil, err
	}
	var set imap.SeqSet
	for _, item := range ranges {
		set.AddRange(item[0], item[1])
	}
	return set, nil
}

func imapParseUIDSet(value string) (imap.UIDSet, error) {
	ranges, err := imapParseNumSet(value)
	if err != nil {
		return nil, err
	}
	var set imap.UIDSet
	for _, item := range ranges {
		set.AddRange(imap.UID(item[0]), imap.UID(item[1]))
	}
	return set, nil
}

func imapTokenizeSearch(value string) ([]string, error) {
	tokens := make([]string, 0)
	for i := 0; i < len(value); {
		switch value[i] {
		case ' ', '\t', '\r', '\n':
			i++
		case '(', ')':
			tokens = append(tokens, value[i:i+1])
			i++
		case '"':
			i++
			var token strings.Builder
			closed := false
			for i < len(value) {
				if value[i] == '\\' && i+1 < len(value) {
					token.WriteByte(value[i+1])
					i += 2
					continue
				}
				if value[i] == '"' {
					i++
					closed = true
					break
				}
				token.WriteByte(value[i])
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated quoted string in IMAP search")
			}
			tokens = append(tokens, token.String())
		default:
			start := i
			for i < len(value) && value[i] != ' ' && value[i] != '\t' && value[i] != '\r' && value[i] != '\n' && value[i] != '(' && value[i] != ')' {
				i++
			}
			tokens = append(tokens, value[start:i])
		}
	}
	return tokens, nil
}

type imapSearchParser struct {
	tokens []string
	pos    int
}

func (p *imapSearchParser) nextValue() (string, error) {
	if p.pos >= len(p.tokens) {
		return "", fmt.Errorf("missing value in IMAP search")
	}
	value := p.tokens[p.pos]
	p.pos++
	if value == "(" || value == ")" {
		return "", fmt.Errorf("unexpected %q in IMAP search", value)
	}
	return value, nil
}

func (p *imapSearchParser) parseList(stopAtParen bool) (*imap.SearchCriteria, error) {
	criteria := &imap.SearchCriteria{}
	for p.pos < len(p.tokens) {
		if p.tokens[p.pos] == ")" {
			if stopAtParen {
				return criteria, nil
			}
			return nil, fmt.Errorf("unexpected ')' in IMAP search")
		}
		item, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		criteria.And(item)
	}
	if stopAtParen {
		return nil, fmt.Errorf("missing ')' in IMAP search")
	}
	return criteria, nil
}

func (p *imapSearchParser) parseKey() (*imap.SearchCriteria, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("missing search key")
	}
	token := p.tokens[p.pos]
	p.pos++

	if token == "(" {
		criteria, err := p.parseList(true)
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos] != ")" {
			return nil, fmt.Errorf("missing ')' in IMAP search")
		}
		p.pos++
		return criteria, nil
	}
	if token == ")" {
		return nil, fmt.Errorf("unexpected ')' in IMAP search")
	}

	key := strings.ToUpper(token)
	switch key {
	case "ALL":
		return &imap.SearchCriteria{}, nil
	case "NOT":
		item, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{Not: []imap.SearchCriteria{*item}}, nil
	case "OR":
		left, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		right, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{Or: [][2]imap.SearchCriteria{{*left, *right}}}, nil
	case "ANSWERED", "DELETED", "DRAFT", "FLAGGED", "RECENT", "SEEN":
		return &imap.SearchCriteria{Flag: []imap.Flag{imapSearchFlags[key]}}, nil
	case "UNANSWERED", "UNDELETED", "UNDRAFT", "UNFLAGGED", "UNSEEN":
		return &imap.SearchCriteria{NotFlag: []imap.Flag{imapSearchFlags[strings.TrimPrefix(key, "UN")]}}, nil
	case "NEW":
		return &imap.SearchCriteria{
			Flag:    []imap.Flag{imap.Flag("\\Recent")},
			NotFlag: []imap.Flag{imap.FlagSeen},
		}, nil
	case "OLD":
		return &imap.SearchCriteria{NotFlag: []imap.Flag{imap.Flag("\\Recent")}}, nil
	case "KEYWORD", "UNKEYWORD":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		if key == "KEYWORD" {
			return &imap.SearchCriteria{Flag: []imap.Flag{imap.Flag(value)}}, nil
		}
		return &imap.SearchCriteria{NotFlag: []imap.Flag{imap.Flag(value)}}, nil
	case "BCC", "CC", "FROM", "SUBJECT", "TO":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		field := map[string]string{"BCC": "Bcc", "CC": "Cc", "FROM": "From", "SUBJECT": "Subject", "TO": "To"}[key]
		return &imap.SearchCriteria{Header: []imap.SearchCriteriaHeaderField{{Key: field, Value: value}}}, nil
	case "HEADER":
		field, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{Header: []imap.SearchCriteriaHeaderField{{Key: field, Value: value}}}, nil
	case "SINCE", "BEFORE", "ON", "SENTSINCE", "SENTBEFORE", "SENTON":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		date, err := imapParseDate(value)
		if err != nil {
			return nil, err
		}
		criteria := &imap.SearchCriteria{}
		switch key {
		case "SINCE":
			criteria.Since = date
		case "BEFORE":
			criteria.Before = date
		case "ON":
			criteria.Since = date
			criteria.Before = date.Add(24 * time.Hour)
		case "SENTSINCE":
			criteria.SentSince = date
		case "SENTBEFORE":
			criteria.SentBefore = date
		case "SENTON":
			criteria.SentSince = date
			criteria.SentBefore = date.Add(24 * time.Hour)
		}
		return criteria, nil
	case "BODY":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{Body: []string{value}}, nil
	case "TEXT":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{Text: []string{value}}, nil
	case "LARGER", "SMALLER":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		size, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size %q in IMAP search", value)
		}
		if key == "LARGER" {
			return &imap.SearchCriteria{Larger: size}, nil
		}
		return &imap.SearchCriteria{Smaller: size}, nil
	case "UID":
		value, err := p.nextValue()
		if err != nil {
			return nil, err
		}
		uids, err := imapParseUIDSet(value)
		if err != nil {
			return nil, err
		}
		return &imap.SearchCriteria{UID: []imap.UIDSet{uids}}, nil
	}

	if seqSet, err := imapParseSeqSet(token); err == nil {
		return &imap.SearchCriteria{SeqNum: []imap.SeqSet{seqSet}}, nil
	}
	return nil, fmt.Errorf("unsupported IMAP search key %q", token)
}

func imapParseSearch(value string) (*imap.SearchCriteria, error) {
	tokens, err := imapTokenizeSearch(value)
	if err != nil {
		return nil, err
	}
	parser := &imapSearchParser{tokens: tokens}
	return parser.parseList(false)
}

func imapFetchHasUIDs(messages []*imapclient.FetchMessageBuffer) bool {
	for _, message := range messages {
		if message.UID != 0 {
			return true
		}
	}
	return false
}

func imapFetchOverviews(client *imapclient.Client, uids []imap.UID) ([]env.Object, error) {
	if len(uids) == 0 {
		return []env.Object{}, nil
	}
	messages, err := client.Fetch(imapUIDSet(uids), &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
		RFC822Size:   true,
		UID:          true,
	}).Collect()
	if err != nil {
		return nil, err
	}

	byUID := make(map[imap.UID]env.Object, len(messages))
	for _, message := range messages {
		byUID[message.UID] = imapOverviewDict(message)
	}

	result := make([]env.Object, 0, len(uids))
	for _, uid := range uids {
		if overview, ok := byUID[uid]; ok {
			result = append(result, overview)
		}
	}
	// Fall back to server order if the server did not echo UIDs at all.
	if len(result) == 0 && !imapFetchHasUIDs(messages) {
		for _, message := range messages {
			result = append(result, imapOverviewDict(message))
		}
	}
	return result, nil
}

func imapFetchEmails(client *imapclient.Client, uids []imap.UID) ([]env.Object, error) {
	if len(uids) == 0 {
		return []env.Object{}, nil
	}
	bodySection := &imap.FetchItemBodySection{Peek: true}
	messages, err := client.Fetch(imapUIDSet(uids), &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
		RFC822Size:   true,
		UID:          true,
		BodySection:  []*imap.FetchItemBodySection{bodySection},
	}).Collect()
	if err != nil {
		return nil, err
	}

	byUID := make(map[imap.UID]env.Object, len(messages))
	for _, message := range messages {
		byUID[message.UID] = imapEmailDict(message, message.FindBodySection(bodySection))
	}

	result := make([]env.Object, 0, len(uids))
	for _, uid := range uids {
		if email, ok := byUID[uid]; ok {
			result = append(result, email)
		}
	}
	if len(result) == 0 && !imapFetchHasUIDs(messages) {
		for _, message := range messages {
			result = append(result, imapEmailDict(message, message.FindBodySection(bodySection)))
		}
	}
	return result, nil
}

func imapFetchRawEmails(client *imapclient.Client, uids []imap.UID) ([]env.Object, error) {
	if len(uids) == 0 {
		return []env.Object{}, nil
	}
	bodySection := &imap.FetchItemBodySection{Peek: true}
	messages, err := client.Fetch(imapUIDSet(uids), &imap.FetchOptions{
		Envelope:     true,
		Flags:        true,
		InternalDate: true,
		RFC822Size:   true,
		UID:          true,
		BodySection:  []*imap.FetchItemBodySection{bodySection},
	}).Collect()
	if err != nil {
		return nil, err
	}

	byUID := make(map[imap.UID]env.Object, len(messages))
	for _, message := range messages {
		dict := imapOverviewDict(message)
		dict.Data["raw"] = *env.NewString(string(message.FindBodySection(bodySection)))
		byUID[message.UID] = dict
	}

	result := make([]env.Object, 0, len(uids))
	for _, uid := range uids {
		if email, ok := byUID[uid]; ok {
			result = append(result, email)
		}
	}
	if len(result) == 0 && !imapFetchHasUIDs(messages) {
		for _, message := range messages {
			dict := imapOverviewDict(message)
			dict.Data["raw"] = *env.NewString(string(message.FindBodySection(bodySection)))
			result = append(result, dict)
		}
	}
	return result, nil
}

func imapMailboxInfo(client *imapclient.Client) (env.Dict, error) {
	mailbox := client.Mailbox()
	if mailbox == nil {
		return env.Dict{}, fmt.Errorf("no folder is selected")
	}
	dict := env.NewDict(make(map[string]any))
	dict.Data["name"] = *env.NewString(mailbox.Name)
	dict.Data["messages"] = *env.NewInteger(int64(mailbox.NumMessages))
	dict.Data["flags"] = imapFlagsBlock(mailbox.Flags)
	dict.Data["permanent-flags"] = imapFlagsBlock(mailbox.PermanentFlags)
	return *dict, nil
}

var Builtins_imap = map[string]*env.Builtin{

	//
	// ##### IMAP client ##### ""
	//
	// Example:
	//  ; Connect using an IMAP URI and list folders
	//  client: Open imap://user:password@imap.gmail.com:993
	//  client .Get-folders |for { .print }
	//
	//  ; The URI can select a mailbox initially
	//  client: Open imaps://user:password@imap.gmail.com/INBOX
	//
	//  ; Search, fetch and update messages
	//  uids: client .Search-emails "UNSEEN SINCE 15-Jan-2023"
	//  emails: client .Get-emails uids
	//  emails |for { email |print email -> "subject" }
	//  uids |for { client .Mark-seen }
	//  client .Close
	//
	// Args:
	// * uri: IMAP URI in the form imap[s]://[user[:pass]@]host[:port][/mailbox]
	// Returns:
	// * Native imap-client object on success
	// * Error on connection/authentication failure
	"imap-uri//Open": {
		Argsn: 1,
		Doc:   "Opens an IMAP client connection. URI format: imap[s]://user[:pass]@host[:port][/mailbox].",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				scheme := ps.Idx.GetWord(uri.Scheme.Index)
				parsed, err := parseIMAPURI(scheme, uri.GetPath())
				if err != nil {
					return evaldo.MakeBuiltinError(ps, err.Error(), "imap-uri//Open")
				}
				return imapNewClient(ps, parsed, "", "", "password", "imap-uri//Open")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "imap-uri//Open")
			}
		},
	},

	// Example:
	//  client: Open\pwd imap://user@imap.gmail.com:993 "app-password"
	// Args:
	// * uri: IMAP URI in the form imap[s]://user@host[:port][/mailbox]
	// * password: Password for LOGIN authentication
	// Returns:
	// * Native imap-client object on success
	"imap-uri//Open\\pwd": {
		Argsn: 2,
		Doc:   "Opens an IMAP client connection with a separate password parameter.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				switch password := arg1.(type) {
				case env.String:
					scheme := ps.Idx.GetWord(uri.Scheme.Index)
					parsed, err := parseIMAPURI(scheme, uri.GetPath())
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "imap-uri//Open\\pwd")
					}
					return imapNewClient(ps, parsed, password.Value, "", "password", "imap-uri//Open\\pwd")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-uri//Open\\pwd")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "imap-uri//Open\\pwd")
			}
		},
	},

	// Example:
	//  client: Open\oauth2 imap://user@gmail.com@imap.gmail.com:993 "ya29.access-token"
	//  ; with OAuth2 the username is taken from the URI and the token is passed separately
	//  client: Open\oauth2 imap://user@gmail.com@imap.gmail.com:993 token
	// Args:
	// * uri: IMAP URI in the form imap[s]://user@host[:port][/mailbox]
	// * token: OAuth2 access token
	// Returns:
	// * Native imap-client object on success
	"imap-uri//Open\\oauth2": {
		Argsn: 2,
		Doc:   "Opens an IMAP client connection with OAuth2 (XOAUTH2/OAUTHBEARER) authentication.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch uri := arg0.(type) {
			case env.Uri:
				switch token := arg1.(type) {
				case env.String:
					scheme := ps.Idx.GetWord(uri.Scheme.Index)
					parsed, err := parseIMAPURI(scheme, uri.GetPath())
					if err != nil {
						return evaldo.MakeBuiltinError(ps, err.Error(), "imap-uri//Open\\oauth2")
					}
					return imapNewClient(ps, parsed, "", token.Value, "oauth2", "imap-uri//Open\\oauth2")
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-uri//Open\\oauth2")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.UriType}, "imap-uri//Open\\oauth2")
			}
		},
	},

	// Example:
	//  client: imap-client "user@gmail.com" "password" "imap.gmail.com" 993
	// Args:
	// * username: String - IMAP username
	// * password: String - IMAP password
	// * server: String - IMAP server hostname (e.g., "imap.gmail.com")
	// * port: Integer - IMAP server port (usually 993 for SSL, 143 for STARTTLS)
	// Returns:
	// * Native imap-client object on success
	"imap-client": {
		Argsn: 4,
		Doc:   "Create an IMAP client connection with username, password, server, and port.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch username := arg0.(type) {
			case env.String:
				switch password := arg1.(type) {
				case env.String:
					switch server := arg2.(type) {
					case env.String:
						switch port := arg3.(type) {
						case env.Integer:
							parsed := &imapURI{
								user:        username.Value,
								password:    password.Value,
								host:        server.Value,
								port:        int(port.Value),
								implicitTLS: int(port.Value) == 993,
							}
							return imapNewClient(ps, parsed, "", "", "password", "imap-client")
						default:
							return evaldo.MakeArgError(ps, 4, []env.Type{env.IntegerType}, "imap-client")
						}
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "imap-client")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "imap-client")
			}
		},
	},

	// Example:
	//  client: imap-client\oauth2 "user@gmail.com" "ya29.token" "imap.gmail.com" 993
	// Args:
	// * email: String - Email address for OAuth2 authentication
	// * access_token: String - OAuth2 access token
	// * server: String - IMAP server hostname
	// * port: Integer - IMAP server port
	// Returns:
	// * Native imap-client object on success
	"imap-client\\oauth2": {
		Argsn: 4,
		Doc:   "Create an IMAP client connection with OAuth2 authentication (email, access_token, server, port).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.String:
				switch token := arg1.(type) {
				case env.String:
					switch server := arg2.(type) {
					case env.String:
						switch port := arg3.(type) {
						case env.Integer:
							parsed := &imapURI{
								user:        email.Value,
								host:        server.Value,
								port:        int(port.Value),
								implicitTLS: int(port.Value) == 993,
							}
							return imapNewClient(ps, parsed, "", token.Value, "oauth2", "imap-client\\oauth2")
						default:
							return evaldo.MakeArgError(ps, 4, []env.Type{env.IntegerType}, "imap-client\\oauth2")
						}
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "imap-client\\oauth2")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client\\oauth2")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.StringType}, "imap-client\\oauth2")
			}
		},
	},

	// Example:
	//  client .Select-folder "INBOX"
	// Args:
	// * client: Native imap-client object
	// * folder: String - Folder name (e.g., "INBOX", "Sent", "Drafts")
	// Returns:
	// * Native imap-client object (for method chaining)
	"imap-client//Select-folder": {
		Argsn: 2,
		Doc:   "Selects a folder for IMAP operations (e.g., 'INBOX').",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Select-folder")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Select-folder")
			}
			switch folder := arg1.(type) {
			case env.String:
				if _, err := client.Select(folder.Value, nil).Wait(); err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to select folder: %s", err), "imap-client//Select-folder")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client//Select-folder")
			}
		},
	},

	// Example:
	//  folders: client .Get-folders
	// Args:
	// * client: Native imap-client object
	// Returns:
	// * Block of strings containing folder names
	"imap-client//Get-folders": {
		Argsn: 1,
		Doc:   "Gets the list of available folders.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Get-folders")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Get-folders")
			}
			mailboxes, err := client.List("", "*", nil).Collect()
			if err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to get folders: %s", err), "imap-client//Get-folders")
			}
			result := make([]env.Object, len(mailboxes))
			for i, mailbox := range mailboxes {
				result[i] = *env.NewString(mailbox.Mailbox)
			}
			return *env.NewBlock(*env.NewTSeries(result))
		},
	},

	// Example:
	//  client .Select-folder "INBOX"
	//  info: client .Mailbox-info
	//  print join { "Messages: " string info -> "messages" }
	// Args:
	// * client: Native imap-client object
	// Returns:
	// * Dictionary with name, messages, flags and permanent-flags for the selected folder
	"imap-client//Mailbox-info": {
		Argsn: 1,
		Doc:   "Gets information about the currently selected folder.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Mailbox-info")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Mailbox-info")
			}
			info, err := imapMailboxInfo(client)
			if err != nil {
				return evaldo.MakeBuiltinError(ps, err.Error(), "imap-client//Mailbox-info")
			}
			return info
		},
	},

	// Example:
	//  uids: client .Search-emails "UNSEEN SINCE 15-Jan-2023"
	//  uids: client .Search-emails "FROM \"sender@example.com\" SUBJECT \"invoice\""
	//  uids: client .Search-emails "NOT DELETED"
	// Args:
	// * client: Native imap-client object
	// * search_criteria: String - IMAP search criteria
	// Returns:
	// * Block of integers containing matching email UIDs
	"imap-client//Search-emails": {
		Argsn: 2,
		Doc:   "Searches for emails using IMAP search criteria and returns matching UIDs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Search-emails")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Search-emails")
			}
			switch criteriaValue := arg1.(type) {
			case env.String:
				criteria, err := imapParseSearch(criteriaValue.Value)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("invalid search criteria: %s", err), "imap-client//Search-emails")
				}
				data, err := client.UIDSearch(criteria, nil).Wait()
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("search failed: %s", err), "imap-client//Search-emails")
				}
				uids := data.AllUIDs()
				result := make([]env.Object, len(uids))
				for i, uid := range uids {
					result[i] = *env.NewInteger(int64(uid))
				}
				return *env.NewBlock(*env.NewTSeries(result))
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client//Search-emails")
			}
		},
	},

	// Example:
	//  overviews: client .Get-overviews { 123 124 125 }
	//  overviews |for { -> "subject" |print }
	// Args:
	// * client: Native imap-client object
	// * uids: Block of integers containing email UIDs
	// Returns:
	// * Block of dictionaries with uid, subject, from, to, cc, bcc, date, received, message-id, size and flags
	"imap-client//Get-overviews": {
		Argsn: 2,
		Doc:   "Gets email overviews (headers only, fast) for a block of UIDs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Get-overviews")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Get-overviews")
			}
			switch uidBlock := arg1.(type) {
			case env.Block:
				uids, argErr := imapUIDsFromBlock(ps, uidBlock, "imap-client//Get-overviews")
				if argErr != nil {
					return argErr
				}
				overviews, err := imapFetchOverviews(client, uids)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to get overviews: %s", err), "imap-client//Get-overviews")
				}
				return *env.NewBlock(*env.NewTSeries(overviews))
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "imap-client//Get-overviews")
			}
		},
	},

	// Example:
	//  emails: client .Get-emails { 123 124 }
	//  emails |for { email |print email -> "subject" }
	// Args:
	// * client: Native imap-client object
	// * uids: Block of integers containing email UIDs
	// Returns:
	// * Block of dictionaries with uid, subject, from, to, cc, bcc, sender, reply-to,
	//   date, received, message-id, size, text, html, flags and attachments
	"imap-client//Get-emails": {
		Argsn: 2,
		Doc:   "Gets full emails with bodies and attachments (slower) for a block of UIDs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Get-emails")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Get-emails")
			}
			switch uidBlock := arg1.(type) {
			case env.Block:
				uids, argErr := imapUIDsFromBlock(ps, uidBlock, "imap-client//Get-emails")
				if argErr != nil {
					return argErr
				}
				emails, err := imapFetchEmails(client, uids)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to get emails: %s", err), "imap-client//Get-emails")
				}
				return *env.NewBlock(*env.NewTSeries(emails))
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "imap-client//Get-emails")
			}
		},
	},

	// Example:
	//  uids: client .Search-emails "ALL"
	//  raw-emails: client .Get-raw-emails uids
	//  raw-emails |for { ::email |print email -> "raw" }
	// Args:
	// * client: Native imap-client object
	// * uids: Block of integers containing email UIDs
	// Returns:
	// * Block of dictionaries with uid, subject, from, to, cc, bcc, date, received,
	//   message-id, size, flags and raw (the complete RFC 822 message)
	"imap-client//Get-raw-emails": {
		Argsn: 2,
		Doc:   "Gets complete raw RFC 822 messages for a block of UIDs.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Get-raw-emails")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Get-raw-emails")
			}
			switch uidBlock := arg1.(type) {
			case env.Block:
				uids, argErr := imapUIDsFromBlock(ps, uidBlock, "imap-client//Get-raw-emails")
				if argErr != nil {
					return argErr
				}
				emails, err := imapFetchRawEmails(client, uids)
				if err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to get raw emails: %s", err), "imap-client//Get-raw-emails")
				}
				return *env.NewBlock(*env.NewTSeries(emails))
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.BlockType}, "imap-client//Get-raw-emails")
			}
		},
	},

	// Example:
	//  client .Mark-seen 123
	// Args:
	// * client: Native imap-client object
	// * uid: Integer - Email UID
	// Returns:
	// * Native imap-client object (for method chaining)
	"imap-client//Mark-seen": {
		Argsn: 2,
		Doc:   "Marks an email as read/seen by UID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Mark-seen")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Mark-seen")
			}
			switch uid := arg1.(type) {
			case env.Integer:
				cmd := client.Store(imap.UIDSetNum(imap.UID(uid.Value)), &imap.StoreFlags{
					Op:     imap.StoreFlagsAdd,
					Silent: true,
					Flags:  []imap.Flag{imap.FlagSeen},
				}, nil)
				if err := cmd.Close(); err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to mark email as seen: %s", err), "imap-client//Mark-seen")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//Mark-seen")
			}
		},
	},

	// Example:
	//  client .Move-email 123 "Archive"
	// Args:
	// * client: Native imap-client object
	// * uid: Integer - Email UID
	// * folder: String - Destination folder
	// Returns:
	// * Native imap-client object (for method chaining)
	"imap-client//Move-email": {
		Argsn: 3,
		Doc:   "Moves an email to another folder by UID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Move-email")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Move-email")
			}
			switch uid := arg1.(type) {
			case env.Integer:
				switch folder := arg2.(type) {
				case env.String:
					if _, err := client.Move(imap.UIDSetNum(imap.UID(uid.Value)), folder.Value).Wait(); err != nil {
						return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to move email: %s", err), "imap-client//Move-email")
					}
					return arg0
				default:
					return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "imap-client//Move-email")
				}
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//Move-email")
			}
		},
	},

	// Example:
	//  client .Delete-email 123
	//  client .Expunge
	// Args:
	// * client: Native imap-client object
	// * uid: Integer - Email UID
	// Returns:
	// * Native imap-client object (for method chaining)
	"imap-client//Delete-email": {
		Argsn: 2,
		Doc:   "Marks an email for deletion by UID (use expunge to permanently delete).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Delete-email")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Delete-email")
			}
			switch uid := arg1.(type) {
			case env.Integer:
				cmd := client.Store(imap.UIDSetNum(imap.UID(uid.Value)), &imap.StoreFlags{
					Op:     imap.StoreFlagsAdd,
					Silent: true,
					Flags:  []imap.Flag{imap.FlagDeleted},
				}, nil)
				if err := cmd.Close(); err != nil {
					return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to delete email: %s", err), "imap-client//Delete-email")
				}
				return arg0
			default:
				return evaldo.MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//Delete-email")
			}
		},
	},

	// Example:
	//  client .Expunge
	// Args:
	// * client: Native imap-client object
	// Returns:
	// * Native imap-client object (for method chaining)
	"imap-client//Expunge": {
		Argsn: 1,
		Doc:   "Permanently removes emails marked for deletion.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Expunge")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Expunge")
			}
			if err := client.Expunge().Close(); err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to expunge: %s", err), "imap-client//Expunge")
			}
			return arg0
		},
	},

	// Example:
	//  client .Close
	// Args:
	// * client: Native imap-client object
	// Returns:
	// * Void
	"imap-client//Close": {
		Argsn: 1,
		Doc:   "Logs out and closes the IMAP client connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			client, ok := imapClientFrom(arg0)
			if !ok {
				if _, native := arg0.(env.Native); native {
					return evaldo.MakeBuiltinError(ps, "Expected imap-client", "imap-client//Close")
				}
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//Close")
			}
			_ = client.Logout().Wait()
			if err := client.Close(); err != nil {
				return evaldo.MakeBuiltinError(ps, fmt.Sprintf("failed to close connection: %s", err), "imap-client//Close")
			}
			return env.Void{}
		},
	},
}

func init() {
	// imap and imaps use the same implementation; the scheme determines whether
	// implicit TLS is used by default.
	Builtins_imap["imaps-uri//Open"] = Builtins_imap["imap-uri//Open"]
	Builtins_imap["imaps-uri//Open\\pwd"] = Builtins_imap["imap-uri//Open\\pwd"]
	Builtins_imap["imaps-uri//Open\\oauth2"] = Builtins_imap["imap-uri//Open\\oauth2"]
}
