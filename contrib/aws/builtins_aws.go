//go:build b_aws
// +build b_aws

package aws

import (
	"bytes"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"fmt"

	"github.com/go-gomail/gomail"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/client"
	//"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go/service/ses"
)

var Builtins_aws = map[string]*env.Builtin{

	"new-aws-session": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch region := arg0.(type) {
			case env.String:
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region.Value))
				fmt.Println(cfg)
				fmt.Println(err)
				if err != nil {
					return evaldo.MakeError(ps, "Error creating new AWS session: "+err.Error())
				}
				return *env.NewNative(ps.Idx, cfg, "aws-session")
			default:
				return evaldo.MakeError(ps, "A1 not String")
			}
		},
	},

	"aws-session//open-ses": {
		Argsn: 1,
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch cfg := arg0.(type) {
			case env.Native:

				svc := ses.NewFromConfig(cfg.Value.(aws.Config))
				//svc := ses.New(sess.Value.(client.ConfigProvider))
				fmt.Println(svc)
				return *env.NewNative(ps.Idx, svc, "aws-ses-session")
			default:
				return evaldo.MakeError(ps, "Arg 1 not native.")
			}
		},
	},

	"aws-ses-session//send-raw": {
		Argsn: 4,
		Doc:   "[ ses-session* gomail-message from-email recipients ]",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch svc := arg0.(type) {
			case env.Native:
				switch msg := arg1.(type) { // gomail-message
				case env.Native:
					switch recps := arg2.(type) { // recipients
					case env.Block:
						switch fromMl := arg3.(type) { // recipients
						case env.Email:

							var recipients []string
							for rec := range recps.Series.S {
								switch rec2 := recps.Series.S[rec].(type) {
								case env.String:
									recipients = append(recipients, rec2.Value)
								case env.Email:
									recipients = append(recipients, rec2.Address)
								}
							}

							fmt.Println(recipients)

							fmt.Println(msg)

							fromEmail := fromMl.Address

							fmt.Println(recipients)

							var emailRaw bytes.Buffer
							msg.Value.(*gomail.Message).WriteTo(&emailRaw)

							// create new raw message
							message := types.RawMessage{Data: emailRaw.Bytes()}

							fmt.Println("111")
							input := &ses.SendRawEmailInput{Source: &fromEmail, Destinations: recipients, RawMessage: &message}

							fmt.Println("222")
							// send raw email
							_, err := svc.Value.(*ses.Client).SendRawEmail(context.TODO(), input)
							if err != nil {
								return evaldo.MakeError(ps, err.Error())
							}
							fmt.Println("SHOULD SEND")
						default:
							return evaldo.MakeError(ps, "A4 not String")

						}
					default:
						return evaldo.MakeError(ps, "A3 not Block")
					}
				default:
					return evaldo.MakeError(ps, "A2 not Native")
				}
			default:
				return evaldo.MakeError(ps, "A1 not Native")
			}
			return nil
		},
	},
}
