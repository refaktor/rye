//go:build !no_imap
// +build !no_imap

package evaldo

import (
	"fmt"
	"strings"
	"time"

	"github.com/refaktor/rye/env"

	imap "github.com/BrianLeishman/go-imap"
)

var Builtins_imap = map[string]*env.Builtin{

	//
	// ##### Imap client ##### ""
	//
	// Example:
	//  ; Connect to IMAP server and list folders
	//  client: imap-client "user@gmail.com" "password" "imap.gmail.com" 993
	//  folders: client .Get-folders
	//  folders |for { .print }
	//
	//  ; Select inbox and search for unread emails
	//  client .Select-folder "INBOX"
	//  uids: client .Search-emails "UNSEEN"
	//
	//  ; Get email overviews (headers only - fast)
	//  overviews: client .Get-overviews uids
	//  overviews |for { -> "subject" |print }
	//
	//  ; Get full email content
	//  emails: client .Get-emails uids
	//  emails |for { email |
	//    print email -> "subject"
	//    print email -> "text"
	//  }
	//
	//  ; Mark as read and close
	//  uids |for { client .Mark-seen }
	//  client .Close
	//
	// Creates a new IMAP client connection using username/password authentication
	// Args:
	// * username: String - IMAP username
	// * password: String - IMAP password
	// * server: String - IMAP server hostname (e.g., "imap.gmail.com")
	// * port: Integer - IMAP server port (usually 993 for SSL, 143 for plain)
	// Returns:
	// * Native imap-client object on success
	// * Error on connection failure
	// Tests:
	// ; client: new-imap-client "user@gmail.com" "password" "imap.gmail.com" 993
	// ; equal { client .type? } 'native
	"imap-client": {
		Argsn: 4,
		Doc:   "Create new IMAP client connection with username, password, server, and port.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch username := arg0.(type) {
			case env.String:
				switch password := arg1.(type) {
				case env.String:
					switch server := arg2.(type) {
					case env.String:
						switch port := arg3.(type) {
						case env.Integer:
							client, err := imap.New(username.Value, password.Value, server.Value, int(port.Value))
							if err != nil {
								ps.FailureFlag = true
								return env.NewError(fmt.Sprintf("IMAP connection failed: %s", err.Error()))
							}
							return *env.NewNative(ps.Idx, client, "imap-client")
						default:
							return MakeArgError(ps, 4, []env.Type{env.IntegerType}, "new-imap-client")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "new-imap-client")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "new-imap-client")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "new-imap-client")
			}
		},
	},

	// Creates a new IMAP client connection using OAuth2 authentication
	// Args:
	// * email: String - Email address for OAuth2 authentication
	// * access_token: String - OAuth2 access token
	// * server: String - IMAP server hostname (e.g., "imap.gmail.com")
	// * port: Integer - IMAP server port (usually 993 for SSL)
	// Returns:
	// * Native imap-client object on success
	// * Error on connection failure
	// Tests:
	// ; oauth-client: new-imap-client-oauth2 "user@gmail.com" "ya29.access_token" "imap.gmail.com" 993
	// ; equal { oauth-client .type? } 'native
	"imap-client\\oauth2": {
		Argsn: 4,
		Doc:   "Create new IMAP client connection with OAuth2 authentication (email, access_token, server, port).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch email := arg0.(type) {
			case env.String:
				switch token := arg1.(type) {
				case env.String:
					switch server := arg2.(type) {
					case env.String:
						switch port := arg3.(type) {
						case env.Integer:
							client, err := imap.NewWithOAuth2(email.Value, token.Value, server.Value, int(port.Value))
							if err != nil {
								ps.FailureFlag = true
								return env.NewError(fmt.Sprintf("IMAP OAuth2 connection failed: %s", err.Error()))
							}
							return *env.NewNative(ps.Idx, client, "imap-client")
						default:
							return MakeArgError(ps, 4, []env.Type{env.IntegerType}, "new-imap-client-oauth2")
						}
					default:
						return MakeArgError(ps, 3, []env.Type{env.StringType}, "new-imap-client-oauth2")
					}
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "new-imap-client-oauth2")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "new-imap-client-oauth2")
			}
		},
	},

	// Selects a folder/mailbox for IMAP operations
	// Args:
	// * client: Native imap-client object
	// * folder: String - Folder name (e.g., "INBOX", "Sent", "Drafts")
	// Returns:
	// * Native imap-client object (for method chaining)
	// * Error if folder selection fails
	// Tests:
	// ; client .Select-folder "INBOX"
	// ; client .Select-folder "Sent Items"
	"imap-client//Select-folder": {
		Argsn: 2,
		Doc:   "Select a folder for operations (e.g., 'INBOX').",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch folder := arg1.(type) {
					case env.String:
						err := imapClient.SelectFolder(folder.Value)
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Failed to select folder: %s", err.Error()))
						}
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client//select-folder")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//select-folder")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//select-folder")
			}
		},
	},

	// Retrieves list of available folders/mailboxes from the IMAP server
	// Args:
	// * client: Native imap-client object
	// Returns:
	// * Block of strings containing folder names (e.g., ["INBOX", "Sent", "Drafts"])
	// * Error if retrieval fails
	// Tests:
	// ; folders: client .Get-folders
	// ; equal { folders .length } 3
	"imap-client//Get-folders": {
		Argsn: 1,
		Doc:   "Get list of available folders.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					folders, err := imapClient.GetFolders()
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(fmt.Sprintf("Failed to get folders: %s", err.Error()))
					}

					result := make([]env.Object, len(folders))
					for i, folder := range folders {
						result[i] = env.String{Value: folder}
					}
					return *env.NewBlock(*env.NewTSeries(result))
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//get-folders")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//get-folders")
			}
		},
	},

	// Searches for emails using IMAP search criteria and returns matching UIDs
	// Args:
	// * client: Native imap-client object
	// * search_criteria: String - IMAP search criteria (e.g., "UNSEEN", "FROM \"user@domain.com\"", "SUBJECT \"test\"", "SINCE \"01-Jan-2023\"")
	// Returns:
	// * Block of integers containing matching email UIDs
	// * Error if search fails
	// Tests:
	// ; unseen_uids: client .Search-emails "UNSEEN"
	// ; from_uids: client .Search-emails "FROM \"sender@example.com\""
	// ; subject_uids: client .Search-emails "SUBJECT \"Important\""
	"imap-client//Search-emails": {
		Argsn: 2,
		Doc:   "Search for emails using IMAP search criteria (e.g., 'UNSEEN', 'FROM \"user@domain.com\"', 'SUBJECT \"test\"').",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch searchCriteria := arg1.(type) {
					case env.String:
						uids, err := imapClient.GetUIDs(searchCriteria.Value)
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Search failed: %s", err.Error()))
						}

						result := make([]env.Object, len(uids))
						for i, uid := range uids {
							result[i] = env.Integer{Value: int64(uid)}
						}
						return *env.NewBlock(*env.NewTSeries(result))
					default:
						return MakeArgError(ps, 2, []env.Type{env.StringType}, "imap-client//search-emails")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//search-emails")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//search-emails")
			}
		},
	},

	// Retrieves email overviews (headers only) for given UIDs - fast operation for listing emails
	// Args:
	// * client: Native imap-client object
	// * uids: Block of integers containing email UIDs to fetch overviews for
	// Returns:
	// * Block of dictionaries with email overview information (uid, subject, from, to, date, size, flags)
	// * Error if retrieval fails
	// Tests:
	// ; overviews: client .Get-overviews { 123 124 125 }
	// ; equal { overviews .length } 3
	// ; equal { overviews -> 0 -> "subject" } "Test Email"
	"imap-client//Get-overviews": {
		Argsn: 2,
		Doc:   "Get email overviews (headers only, fast) for given UIDs block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch uidsBlock := arg1.(type) {
					case env.Block:
						var uids []int
						for _, uidObj := range uidsBlock.Series.S {
							switch uid := uidObj.(type) {
							case env.Integer:
								uids = append(uids, int(uid.Value))
							default:
								return MakeBuiltinError(ps, "All UIDs must be integers", "imap-client//get-overviews")
							}
						}

						if len(uids) == 0 {
							return *env.NewBlock(*env.NewTSeries([]env.Object{}))
						}

						overviews, err := imapClient.GetOverviews(uids...)
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Failed to get overviews: %s", err.Error()))
						}

						result := make([]env.Object, 0, len(overviews))
						for uid, email := range overviews {
							emailDict := env.NewDict(make(map[string]any))
							emailDict.Data["uid"] = *env.NewInteger(int64(uid))
							emailDict.Data["subject"] = *env.NewString(email.Subject)

							// Convert email addresses to strings inline
							fromStr := ""
							if email.From != nil && len(email.From) > 0 {
								var fromParts []string
								for emailAddr, name := range email.From {
									if name != "" {
										fromParts = append(fromParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										fromParts = append(fromParts, emailAddr)
									}
								}
								fromStr = strings.Join(fromParts, ", ")
							}
							emailDict.Data["from"] = *env.NewString(fromStr)

							toStr := ""
							if email.To != nil && len(email.To) > 0 {
								var toParts []string
								for emailAddr, name := range email.To {
									if name != "" {
										toParts = append(toParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										toParts = append(toParts, emailAddr)
									}
								}
								toStr = strings.Join(toParts, ", ")
							}
							emailDict.Data["to"] = *env.NewString(toStr)

							emailDict.Data["date"] = *env.NewString(email.Sent.Format(time.RFC3339))
							emailDict.Data["size"] = *env.NewInteger(int64(email.Size))

							// Add flags as a block
							flagObjs := make([]env.Object, len(email.Flags))
							for i, flag := range email.Flags {
								flagObjs[i] = *env.NewString(flag)
							}
							emailDict.Data["flags"] = *env.NewBlock(*env.NewTSeries(flagObjs))

							result = append(result, *emailDict)
						}

						return *env.NewBlock(*env.NewTSeries(result))
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "imap-client//get-overviews")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//get-overviews")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//get-overviews")
			}
		},
	},

	// Retrieves full email content including bodies and attachments for given UIDs - slower but complete operation
	// Args:
	// * client: Native imap-client object
	// * uids: Block of integers containing email UIDs to fetch full content for
	// Returns:
	// * Block of dictionaries with complete email information (uid, subject, from, to, cc, bcc, date, received, message-id, size, text, html, flags, attachments)
	// * Error if retrieval fails
	// Tests:
	// ; full_emails: client .Get-emails { 123 124 }
	// ; equal { full_emails -> 0 -> "text" .length > 0 } true
	// ; equal { full_emails -> 0 -> "attachments" .length } 2
	"imap-client//Get-emails": {
		Argsn: 2,
		Doc:   "Get full emails with bodies and attachments (slower) for given UIDs block.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch uidsBlock := arg1.(type) {
					case env.Block:
						var uids []int
						for _, uidObj := range uidsBlock.Series.S {
							switch uid := uidObj.(type) {
							case env.Integer:
								uids = append(uids, int(uid.Value))
							default:
								return MakeBuiltinError(ps, "All UIDs must be integers", "imap-client//get-emails")
							}
						}

						if len(uids) == 0 {
							return *env.NewBlock(*env.NewTSeries([]env.Object{}))
						}

						emails, err := imapClient.GetEmails(uids...)
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Failed to get emails: %s", err.Error()))
						}

						result := make([]env.Object, 0, len(emails))
						for uid, email := range emails {
							emailDict := env.NewDict(make(map[string]any))
							emailDict.Data["uid"] = *env.NewInteger(int64(uid))
							emailDict.Data["subject"] = *env.NewString(email.Subject)

							// Convert email addresses to strings inline
							fromStr := ""
							if email.From != nil && len(email.From) > 0 {
								var fromParts []string
								for emailAddr, name := range email.From {
									if name != "" {
										fromParts = append(fromParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										fromParts = append(fromParts, emailAddr)
									}
								}
								fromStr = strings.Join(fromParts, ", ")
							}
							emailDict.Data["from"] = *env.NewString(fromStr)

							toStr := ""
							if email.To != nil && len(email.To) > 0 {
								var toParts []string
								for emailAddr, name := range email.To {
									if name != "" {
										toParts = append(toParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										toParts = append(toParts, emailAddr)
									}
								}
								toStr = strings.Join(toParts, ", ")
							}
							emailDict.Data["to"] = *env.NewString(toStr)

							ccStr := ""
							if email.CC != nil && len(email.CC) > 0 {
								var ccParts []string
								for emailAddr, name := range email.CC {
									if name != "" {
										ccParts = append(ccParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										ccParts = append(ccParts, emailAddr)
									}
								}
								ccStr = strings.Join(ccParts, ", ")
							}
							emailDict.Data["cc"] = *env.NewString(ccStr)

							bccStr := ""
							if email.BCC != nil && len(email.BCC) > 0 {
								var bccParts []string
								for emailAddr, name := range email.BCC {
									if name != "" {
										bccParts = append(bccParts, fmt.Sprintf("%s <%s>", name, emailAddr))
									} else {
										bccParts = append(bccParts, emailAddr)
									}
								}
								bccStr = strings.Join(bccParts, ", ")
							}
							emailDict.Data["bcc"] = *env.NewString(bccStr)

							emailDict.Data["date"] = *env.NewString(email.Sent.Format(time.RFC3339))
							emailDict.Data["received"] = *env.NewString(email.Received.Format(time.RFC3339))
							emailDict.Data["message-id"] = *env.NewString(email.MessageID)
							emailDict.Data["size"] = *env.NewInteger(int64(email.Size))
							emailDict.Data["text"] = *env.NewString(email.Text)
							emailDict.Data["html"] = *env.NewString(email.HTML)

							// Add flags as a block
							flagObjs := make([]env.Object, len(email.Flags))
							for i, flag := range email.Flags {
								flagObjs[i] = *env.NewString(flag)
							}
							emailDict.Data["flags"] = *env.NewBlock(*env.NewTSeries(flagObjs))

							// Add attachments as a block of dictionaries
							if len(email.Attachments) > 0 {
								attachObjs := make([]env.Object, len(email.Attachments))
								for i, att := range email.Attachments {
									attDict := env.NewDict(make(map[string]any))
									attDict.Data["name"] = *env.NewString(att.Name)
									attDict.Data["mime-type"] = *env.NewString(att.MimeType)
									attDict.Data["size"] = *env.NewInteger(int64(len(att.Content)))
									// Store content as base64 encoded string since there's no binary type
									attDict.Data["content"] = string(att.Content)
									attachObjs[i] = *attDict
								}
								emailDict.Data["attachments"] = *env.NewBlock(*env.NewTSeries(attachObjs))
							} else {
								emailDict.Data["attachments"] = *env.NewBlock(*env.NewTSeries([]env.Object{}))
							}

							result = append(result, *emailDict)
						}

						return *env.NewBlock(*env.NewTSeries(result))
					default:
						return MakeArgError(ps, 2, []env.Type{env.BlockType}, "imap-client//get-emails")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//get-emails")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//get-emails")
			}
		},
	},

	"imap-client//Mark-seen": {
		Argsn: 2,
		Doc:   "Mark an email as read/seen by UID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch uid := arg1.(type) {
					case env.Integer:
						err := imapClient.MarkSeen(int(uid.Value))
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Failed to mark email as seen: %s", err.Error()))
						}
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//mark-seen")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//mark-seen")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//mark-seen")
			}
		},
	},

	"imap-client//Move-email": {
		Argsn: 3,
		Doc:   "Move an email to another folder by UID.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch uid := arg1.(type) {
					case env.Integer:
						switch folder := arg2.(type) {
						case env.String:
							err := imapClient.MoveEmail(int(uid.Value), folder.Value)
							if err != nil {
								ps.FailureFlag = true
								return env.NewError(fmt.Sprintf("Failed to move email: %s", err.Error()))
							}
							return arg0
						default:
							return MakeArgError(ps, 3, []env.Type{env.StringType}, "imap-client//move-email")
						}
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//move-email")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//move-email")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//move-email")
			}
		},
	},

	"imap-client//Delete-email": {
		Argsn: 2,
		Doc:   "Mark an email for deletion by UID (use expunge to permanently delete).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					switch uid := arg1.(type) {
					case env.Integer:
						err := imapClient.DeleteEmail(int(uid.Value))
						if err != nil {
							ps.FailureFlag = true
							return env.NewError(fmt.Sprintf("Failed to delete email: %s", err.Error()))
						}
						return arg0
					default:
						return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "imap-client//delete-email")
					}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//delete-email")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//delete-email")
			}
		},
	},

	"imap-client//Expunge": {
		Argsn: 1,
		Doc:   "Permanently remove emails marked for deletion.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					err := imapClient.Expunge()
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(fmt.Sprintf("Failed to expunge: %s", err.Error()))
					}
					return arg0
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//expunge")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//expunge")
			}
		},
	},

	"imap-client//Close": {
		Argsn: 1,
		Doc:   "Close the IMAP client connection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				if imapClient, ok := client.Value.(*imap.Dialer); ok {
					err := imapClient.Close()
					if err != nil {
						ps.FailureFlag = true
						return env.NewError(fmt.Sprintf("Failed to close connection: %s", err.Error()))
					}
					return env.Void{}
				} else {
					return MakeBuiltinError(ps, "Expected imap-client", "imap-client//close")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.NativeType}, "imap-client//close")
			}
		},
	},
}
