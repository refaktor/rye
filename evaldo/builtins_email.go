//go:build !no_email
// +build !no_email

package evaldo

import (
	"strings"

	"github.com/refaktor/rye/env"

	"github.com/go-gomail/gomail"
)

func __newMessage(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	// Create a new empty gomail message object
	return *env.NewNative(ps.Idx, gomail.NewMessage(), "gomail-message")
}

func __setHeader(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		var fld string
		var val string
		// Extract header value from String or Email object
		switch value := arg2.(type) {
		case env.String:
			val = value.Value
		case env.Email:
			val = value.Address // Use email address from Email type
		default:
			return MakeArgError(ps, 3, []env.Type{env.StringType, env.EmailType}, "gomail-message//Set-header")
		}
		// Extract field name from String or Tagword
		switch field := arg1.(type) {
		case env.String:
			fld = field.Value
		case env.Tagword:
			fld = ps.Idx.GetWord(field.Index) // Convert tagword to string
		default:
			return MakeArgError(ps, 2, []env.Type{env.StringType, env.TagwordType}, "gomail-message//Set-header")
		}
		// Set the header if both field and value are non-empty
		if fld != "" && val != "" {
			mailobj.Value.(*gomail.Message).SetHeader(fld, val)
			return arg0
		} else {
			return MakeBuiltinError(ps, "Not both values were defined.", "gomail-message//Set-header")
		}
	default:
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//Set-header")
	}
}

func __setAddressHeader(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch field := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				switch name := arg3.(type) {
				case env.String:
					// Set address header with both email address and display name
					mailobj.Value.(*gomail.Message).SetAddressHeader(field.Value, value.Value, name.Value)
					return arg0
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 4, []env.Type{env.StringType}, "gomail-message//Set-address-header")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//Set-address-header")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//Set-address-header")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//Set-address-header")
	}
}

func __setBody(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				// Set the email body with specified MIME content type
				mailobj.Value.(*gomail.Message).SetBody(encoding.Value, value.Value)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//Set-body")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//Set-body")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//Set-body")
	}
}

func __attach(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch file := arg1.(type) {
		case env.Uri:
			// Extract file path from URI (remove scheme prefix like "file://")
			ath := strings.Split(file.Path, "://")
			mailobj.Value.(*gomail.Message).Attach(ath[1])
			return arg0
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.UriType}, "gomail-message//Attach")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//Attach")
	}
}

func __addAlternative(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch mailobj := arg0.(type) {
	case env.Native:
		switch encoding := arg1.(type) {
		case env.String:
			switch value := arg2.(type) {
			case env.String:
				// Add alternative content format (e.g., HTML version alongside plain text)
				mailobj.Value.(*gomail.Message).AddAlternative(encoding.Value, value.Value)
				return arg0
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "gomail-message//Add-alternative")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.StringType}, "gomail-message//Add-alternative")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-message//Add-alternative")
	}
}

func __newDialer(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch url := arg0.(type) {
	case env.String:
		switch port := arg1.(type) {
		case env.Integer:
			switch username := arg2.(type) {
			case env.String:
				switch pwd := arg3.(type) {
				case env.String:
					// Create SMTP dialer with server connection details and authentication
					return *env.NewNative(ps.Idx, gomail.NewDialer(url.Value, int(port.Value), username.Value, pwd.Value), "gomail-dialer")
				default:
					ps.FailureFlag = true
					return MakeArgError(ps, 4, []env.Type{env.StringType}, "new-email-dialer")
				}
			default:
				ps.FailureFlag = true
				return MakeArgError(ps, 3, []env.Type{env.StringType}, "new-email-dialer")
			}
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.IntegerType}, "new-email-dialer")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.StringType}, "new-email-dialer")
	}
}

func __dialAndSend(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
	switch dialer := arg0.(type) {
	case env.Native:
		switch message := arg1.(type) {
		case env.Native:
			// Connect to SMTP server and send the email message
			if err := dialer.Value.(*gomail.Dialer).DialAndSend(message.Value.(*gomail.Message)); err != nil {
				ps.FailureFlag = true
				return env.NewError(err.Error())
			}
			return arg0
		default:
			ps.FailureFlag = true
			return MakeArgError(ps, 2, []env.Type{env.NativeType}, "gomail-dialer//Dial-and-send")
		}
	default:
		ps.FailureFlag = true
		return MakeArgError(ps, 1, []env.Type{env.NativeType}, "gomail-dialer//Dial-and-send")
	}
}

var Builtins_email = map[string]*env.Builtin{

	//
	// ##### Email Message ##### "Creating and configuring email messages"
	//
	// Example:
	// msg: email-message
	// equal { msg |type? } 'native
	// Args:
	// * (none)
	// Returns:
	// * native gomail-message object for building email content
	"email-message": {
		Argsn: 0,
		Doc:   "Creates a new empty email message object that can be configured with headers, body, and attachments.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newMessage(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example:
	// msg: email-message
	// msg .Set-header "Subject" "Test Email"
	// msg .Set-header 'to "user@example.com"
	// error { msg .Set-header "Subject" 123 }
	// Args:
	// * message: Native gomail-message object
	// * field: String or Tagword representing header name (e.g., "Subject", 'to, 'from)
	// * value: String or Email containing the header value
	// Returns:
	// * the message object for method chaining
	"gomail-message//Set-header": {
		Argsn: 3,
		Doc:   "Sets a standard email header field such as Subject, To, From, Cc, or Bcc with the specified value.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setHeader(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example :
	// msg: email-message
	// msg .Set-address-header "From" "sender@example.com" "John Doe"
	// msg .Set-address-header "To" "recipient@example.com" "Jane Smith"
	// error { msg .Set-address-header "From" 123 "Name" }
	// Args:
	// * message: Native gomail-message object
	// * field: String header field name (e.g., "From", "To", "Cc", "Bcc")
	// * address: String email address
	// * name: String display name for the email address
	// Returns:
	// * the message object for method chaining
	"gomail-message//Set-address-header": {
		Argsn: 4,
		Doc:   "Sets an email address header with both email address and display name, commonly used for From, To, Cc, and Bcc fields.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setAddressHeader(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example :
	// msg: email-message
	// msg .Set-body "text/plain" "Hello, this is a plain text email."
	// msg .Set-body "text/html" "<h1>Hello</h1><p>This is an HTML email.</p>"
	// error { msg .Set-body "text/plain" 123 }
	// Args:
	// * message: Native gomail-message object
	// * contentType: String MIME content type (e.g., "text/plain", "text/html")
	// * content: String containing the email body content
	// Returns:
	// * the message object for method chaining
	"gomail-message//Set-body": {
		Argsn: 3,
		Doc:   "Sets the main body content of the email with the specified MIME content type, supporting plain text and HTML formats.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __setBody(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example :
	// msg: email-message
	// msg .Attach %file://document.pdf
	// msg .Attach %file://image.jpg
	// error { msg .Attach "not-a-uri" }
	// Args:
	// * message: Native gomail-message object
	// * file: Uri pointing to the file to attach (must use file:// scheme)
	// Returns:
	// * the message object for method chaining
	"gomail-message//Attach": {
		Argsn: 2,
		Doc:   "Attaches a file to the email message using a file URI, making the file available as an email attachment.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __attach(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example :
	// msg: email-message
	// msg .Set-body "text/plain" "Plain text version"
	// msg .Add-alternative "text/html" "<p>HTML version</p>"
	// error { msg .Add-alternative "text/html" 123 }
	// Args:
	// * message: Native gomail-message object
	// * contentType: String MIME content type for the alternative content
	// * content: String containing the alternative body content
	// Returns:
	// * the message object for method chaining
	"gomail-message//Add-alternative": {
		Argsn: 3,
		Doc:   "Adds alternative content to the email (e.g., HTML version alongside plain text), allowing email clients to choose their preferred format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __addAlternative(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	//
	// ##### Email Sending Functions ##### "SMTP configuration and email delivery."
	//

	// Example :
	// dialer: new-email-dialer "smtp.gmail.com" 587 "user@gmail.com" "password"
	// equal { dialer |type? } 'native
	// error { new-email-dialer "smtp.gmail.com" "not-a-port" "user" "pass" }
	// Args:
	// * server: String SMTP server hostname (e.g., "smtp.gmail.com", "mail.example.com")
	// * port: Integer SMTP port number (commonly 25, 465, 587, or 2525)
	// * username: String username for SMTP authentication
	// * password: String password for SMTP authentication
	// Returns:
	// * native gomail-dialer object configured for sending emails
	"new-email-dialer": {
		Argsn: 4,
		Doc:   "Creates a new SMTP dialer configured with server details and authentication credentials for sending emails.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __newDialer(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},

	// Example :
	// dialer: new-email-dialer "smtp.gmail.com" 587 "user@gmail.com" "password"
	// msg: email-message
	// result: dialer .Dial-and-send msg
	// error { "not-dialer" .Dial-and-send msg }
	// Args:
	// * dialer: Native gomail-dialer object configured with SMTP settings
	// * message: Native gomail-message object containing the email to send
	// Returns:
	// * the dialer object on success, or error object if sending fails
	"gomail-dialer//Dial-and-send": {
		Argsn: 2,
		Doc:   "Connects to the SMTP server and sends the specified email message, handling authentication and delivery.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return __dialAndSend(ps, arg0, arg1, arg2, arg3, arg4)
		},
	},
}
