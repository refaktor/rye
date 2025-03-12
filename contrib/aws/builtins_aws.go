//go:build b_aws
// +build b_aws

package aws

import (
	"bytes"
	"io"
	"time"

	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"

	"fmt"

	"github.com/go-gomail/gomail"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

var Builtins_aws = map[string]*env.Builtin{

	//
	// ##### AWS ##### "AWS SDK functions"
	//
	// Tests:
	// equal { new-aws-session "us-east-1" |type? } 'native
	// equal { new-aws-session "us-east-1" |kind? } 'aws-session
	// Args:
	// * region: string containing the AWS region (e.g., "us-east-1")
	// Returns:
	// * native aws-session object
	"new-aws-session": {
		Argsn: 1,
		Doc:   "Creates a new AWS session with the specified region.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch region := arg0.(type) {
			case env.String:
				cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region.Value))
				if err != nil {
					return evaldo.MakeError(ps, "Error creating new AWS session: "+err.Error())
				}
				return *env.NewNative(ps.Idx, cfg, "aws-session")
			default:
				return evaldo.MakeError(ps, "A1 not String")
			}
		},
	},

	// Tests:
	// equal { new-aws-session "us-east-1" |open-s3 |type? } 'native
	// equal { new-aws-session "us-east-1" |open-s3 |kind? } 'aws-s3-client
	// Args:
	// * session: native aws-session object
	// Returns:
	// * native aws-s3-client object
	"aws-session//open-s3": {
		Argsn: 1,
		Doc:   "Creates an S3 client from an AWS session.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch sess := arg0.(type) {
			case env.Native:
				client := s3.NewFromConfig(sess.Value.(aws.Config))
				return *env.NewNative(ps.Idx, client, "aws-s3-client")
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "aws-session//new-s3-client")
			}
		},
	},

	// Tests:
	// equal { new-aws-session "us-east-1" |open-s3 |put-object "bucket" "key" reader "content" |kind? } 'aws-s3-client
	// Args:
	// * client: native aws-s3-client object
	// * bucket: string containing the S3 bucket name
	// * key: string containing the S3 object key
	// * reader: native reader object containing the data to upload
	// Returns:
	// * the aws-s3-client object if successful
	"aws-s3-client//put-object": {
		Argsn: 4,
		Doc:   "Uploads an object to an S3 bucket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch bucket := arg1.(type) {
				case env.String:
					switch key := arg2.(type) {
					case env.String:
						switch r := arg3.(type) {
						case env.Native:
							reader, ok := r.Value.(io.Reader)
							if !ok {
								return evaldo.MakeError(ps, "Reader argument is not a reader")
							}
							readerCopy := reader
							// Read all content into a buffer to get the size, we need it to call PutObject for buffered readers
							// TODO: This is definitely not OK and just temporary!
							// 	     Think about if we want to limit this function to only work with non-buffered readers or files.
							content, err := io.ReadAll(readerCopy)
							if err != nil {
								return evaldo.MakeError(ps, "Error reading reader: "+err.Error())
							}

							ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()
							s3Client := client.Value.(*s3.Client)
							_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
								Bucket:        aws.String(bucket.Value),
								Key:           aws.String(key.Value),
								Body:          bytes.NewReader(content),
								ContentLength: aws.Int64(int64(len(content))),
							})
							if err != nil {
								return evaldo.MakeError(ps, "Error putting object: "+err.Error())
							}
							return arg0
						default:
							return evaldo.MakeArgError(ps, 4, []env.Type{env.NativeType}, "aws-s3-client//put-object")
						}
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "aws-s3-client//put-object")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "aws-s3-client//put-object")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "aws-s3-client//put-object")
			}
		},
	},

	// Tests:
	// equal { new-aws-session "us-east-1" |open-s3 |get-object "bucket" "key" |kind? } 'reader
	// Args:
	// * client: native aws-s3-client object
	// * bucket: string containing the S3 bucket name
	// * key: string containing the S3 object key
	// Returns:
	// * native reader object containing the downloaded data
	"aws-s3-client//get-object": {
		Argsn: 3,
		Doc:   "Downloads an object from an S3 bucket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch bucket := arg1.(type) {
				case env.String:
					switch key := arg2.(type) {
					case env.String:
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						s3Client := client.Value.(*s3.Client)
						output, err := s3Client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket.Value), Key: aws.String(key.Value)})
						if err != nil {
							return evaldo.MakeError(ps, "Error getting object: "+err.Error())
						}
						defer output.Body.Close()
						// TODO: the output.Body is a reader that is bound to the context, we need to copy it to a new reader
						b, err := io.ReadAll(output.Body)
						if err != nil {
							return evaldo.MakeError(ps, "Error reading object: "+err.Error())
						}
						return *env.NewNative(ps.Idx, bytes.NewReader(b), "reader")
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "aws-s3-client//get-object")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "aws-s3-client//get-object")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "aws-s3-client//get-object")
			}
		},
	},

	// Tests:
	// equal { new-aws-session "us-east-1" |open-s3 |delete-object "bucket" "key" |kind? } 'aws-s3-client
	// Args:
	// * client: native aws-s3-client object
	// * bucket: string containing the S3 bucket name
	// * key: string containing the S3 object key
	// Returns:
	// * the aws-s3-client object if successful
	"aws-s3-client//delete-object": {
		Argsn: 3,
		Doc:   "Deletes an object from an S3 bucket.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch client := arg0.(type) {
			case env.Native:
				switch bucket := arg1.(type) {
				case env.String:
					switch key := arg2.(type) {
					case env.String:
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						s3Client := client.Value.(*s3.Client)
						_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket.Value), Key: aws.String(key.Value)})
						if err != nil {
							return evaldo.MakeError(ps, "Error deleting object: "+err.Error())
						}
						return arg0
					default:
						return evaldo.MakeArgError(ps, 3, []env.Type{env.StringType}, "aws-s3-client//delete-object")
					}
				default:
					return evaldo.MakeArgError(ps, 2, []env.Type{env.StringType}, "aws-s3-client//delete-object")
				}
			default:
				return evaldo.MakeArgError(ps, 1, []env.Type{env.NativeType}, "aws-s3-client//delete-object")
			}
		},
	},

	// Tests:
	// equal { new-aws-session "us-east-1" |open-ses |type? } 'native
	// equal { new-aws-session "us-east-1" |open-ses |kind? } 'aws-ses-session
	// Args:
	// * session: native aws-session object
	// Returns:
	// * native aws-ses-session object
	"aws-session//open-ses": {
		Argsn: 1,
		Doc:   "Creates an SES client from an AWS session.",
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

	// Tests:
	// equal { new-aws-session "us-east-1" |open-ses |send-raw message "sender@example.com" [ "recipient@example.com" ] |kind? } 'aws-ses-session
	// Args:
	// * session: native aws-ses-session object
	// * message: native gomail-message object
	// * recipients: block of strings or emails containing recipient addresses
	// * from: email address to send from
	// Returns:
	// * the aws-ses-session object if successful
	"aws-ses-session//send-raw": {
		Argsn: 4,
		Doc:   "Sends a raw email using Amazon SES.",
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
