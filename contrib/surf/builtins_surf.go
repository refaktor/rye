//go:build b_surf
// +build b_surf

package surf

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/refaktor/rye/env"
	"github.com/refaktor/rye/evaldo"
	surf "gopkg.in/headzoo/surf.v1"
)

var Builtins_surf = map[string]*env.Builtin{

	//
	// ##### Surf Browser ##### "Functions for web scraping and browser automation"
	//

	// Tests:
	// browser: surf
	// Args:
	// Returns:
	// * surf-browser - New browser instance for web automation
	"surf": {
		Argsn: 0,
		Doc:   "Creates a new Surf browser instance for web scraping and automation.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			browser := surf.NewBrowser()
			return *env.NewNative(ps.Idx, browser, "surf-browser")
		},
	},

	// Tests:
	// client: surf-client
	// Args:
	// Returns:
	// * surf-client - New surf client for builder pattern
	"surf-client": {
		Argsn: 0,
		Doc:   "Creates a new Surf client instance that can be used with builder pattern.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			// Create a map to store builder configuration
			builderConfig := make(map[string]interface{})
			builderConfig["type"] = "client"
			return *env.NewNative(ps.Idx, builderConfig, "surf-builder")
		},
	},

	// Tests:
	// builder: surf-client .builder
	// Args:
	// * client: Surf client instance
	// Returns:
	// * surf-builder - Builder instance for configuring the client
	"surf-builder//Builder": {
		Argsn: 1,
		Doc:   "Returns a builder for configuring the Surf client with fluent API.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch builder := arg0.(type) {
			case env.Native:
				if config, ok := builder.Value.(map[string]interface{}); ok {
					config["builder_mode"] = true
					return *env.NewNative(ps.Idx, config, "surf-builder")
				}
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			}
		},
	},

	// Tests:
	// impersonate: builder .impersonate
	// Args:
	// * builder: Surf builder instance
	// Returns:
	// * surf-impersonate - Impersonation builder for choosing browser to impersonate
	"surf-builder//Impersonate": {
		Argsn: 1,
		Doc:   "Enables browser impersonation mode in the builder.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch builder := arg0.(type) {
			case env.Native:
				if config, ok := builder.Value.(map[string]interface{}); ok {
					config["impersonate"] = true
					return *env.NewNative(ps.Idx, config, "surf-impersonate")
				}
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			}
		},
	},

	// Tests:
	// chrome-builder: impersonate .chrome
	// Args:
	// * impersonate: Surf impersonate instance
	// Returns:
	// * surf-builder - Builder configured to impersonate Chrome
	"surf-impersonate//Chrome": {
		Argsn: 1,
		Doc:   "Configures the builder to impersonate Chrome browser.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch impersonate := arg0.(type) {
			case env.Native:
				if config, ok := impersonate.Value.(map[string]interface{}); ok {
					config["browser"] = "chrome"
					config["user_agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
					return *env.NewNative(ps.Idx, config, "surf-builder")
				}
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			}
		},
	},

	// Tests:
	// firefox-builder: impersonate .firefox
	// Args:
	// * impersonate: Surf impersonate instance
	// Returns:
	// * surf-builder - Builder configured to impersonate Firefox
	"surf-impersonate//Firefox": {
		Argsn: 1,
		Doc:   "Configures the builder to impersonate Firefox browser.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch impersonate := arg0.(type) {
			case env.Native:
				if config, ok := impersonate.Value.(map[string]interface{}); ok {
					config["browser"] = "firefox"
					config["user_agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0"
					return *env.NewNative(ps.Idx, config, "surf-builder")
				}
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			}
		},
	},

	// Tests:
	// safari-builder: impersonate .safari
	// Args:
	// * impersonate: Surf impersonate instance
	// Returns:
	// * surf-builder - Builder configured to impersonate Safari
	"surf-impersonate//Safari": {
		Argsn: 1,
		Doc:   "Configures the builder to impersonate Safari browser.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch impersonate := arg0.(type) {
			case env.Native:
				if config, ok := impersonate.Value.(map[string]interface{}); ok {
					config["browser"] = "safari"
					config["user_agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15"
					return *env.NewNative(ps.Idx, config, "surf-builder")
				}
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf impersonate instance.")
			}
		},
	},

	// Tests:
	// session-builder: builder .session
	// Args:
	// * builder: Surf builder instance
	// Returns:
	// * surf-builder - Builder with session management enabled
	"surf-builder//Session": {
		Argsn: 1,
		Doc:   "Enables session management in the builder (cookies, state persistence).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch builder := arg0.(type) {
			case env.Native:
				if config, ok := builder.Value.(map[string]interface{}); ok {
					config["session"] = true
					return *env.NewNative(ps.Idx, config, "surf-builder")
				}
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			}
		},
	},

	// Tests:
	// browser: builder .build
	// Args:
	// * builder: Surf builder instance
	// Returns:
	// * surf-browser - Final browser instance with all configurations applied
	"surf-builder//Build": {
		Argsn: 1,
		Doc:   "Builds and returns the final configured Surf browser instance.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch builder := arg0.(type) {
			case env.Native:
				if config, ok := builder.Value.(map[string]interface{}); ok {
					// Create a new browser with the configuration
					browser := surf.NewBrowser()

					// Apply user agent if configured
					if ua, exists := config["user_agent"]; exists {
						if uaStr, ok := ua.(string); ok {
							browser.SetUserAgent(uaStr)
						}
					}

					return *env.NewNative(ps.Idx, browser, "surf-browser")
				}
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf builder.")
			}
		},
	},

	//
	// ##### Navigation ##### "Functions for browser navigation and page control"
	//

	// Tests:
	// browser .open "https://example.com"
	// Args:
	// * browser: Surf browser instance
	// * url: String - URL to navigate to
	// Returns:
	// * integer - 1 on success
	"surf-browser//Open": {
		Argsn: 2,
		Doc:   "Opens a URL in the browser and navigates to the page.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch url := arg1.(type) {
					case env.String:
						// Use reflection/interface to call Open method
						if opener, ok := bow.(interface{ Open(string) error }); ok {
							err := opener.Open(url.Value)
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Failed to open URL: %s", err.Error()))
							}
							return arg0
						}
						return evaldo.MakeError(ps, "Browser does not support Open method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (URL).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// title: browser .title
	// Args:
	// * browser: Surf browser instance
	// Returns:
	// * string - Current page title
	"surf-browser//Title": {
		Argsn: 1,
		Doc:   "Gets the current page title.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					if titler, ok := bow.(interface{ Title() string }); ok {
						title := titler.Title()
						return *env.NewString(title)
					}
					return evaldo.MakeError(ps, "Browser does not support Title method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// browser .click "button#submit"
	// browser .click ".login-link"
	// Args:
	// * browser: Surf browser instance
	// * selector: String - CSS selector of element to click
	// Returns:
	// * integer - 1 on success
	"surf-browser//Click": {
		Argsn: 2,
		Doc:   "Clicks an element specified by CSS selector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch selector := arg1.(type) {
					case env.String:
						if clicker, ok := bow.(interface{ Click(string) error }); ok {
							err := clicker.Click(selector.Value)
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Failed to click element: %s", err.Error()))
							}
							return *env.NewInteger(1)
						}
						return evaldo.MakeError(ps, "Browser does not support Click method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (CSS selector).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// browser .back
	// Args:
	// * browser: Surf browser instance
	// Returns:
	// * integer - 1 on success
	"surf-browser//Back": {
		Argsn: 1,
		Doc:   "Navigates back to the previous page.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					if backer, ok := bow.(interface{ Back() error }); ok {
						err := backer.Back()
						if err != nil {
							return evaldo.MakeError(ps, fmt.Sprintf("Failed to go back: %s", err.Error()))
						}
						return *env.NewInteger(1)
					}
					return evaldo.MakeError(ps, "Browser does not support Back method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// browser .forward
	// Args:
	// * browser: Surf browser instance
	// Returns:
	// * integer - 1 on success
	"surf-browser//Forward": {
		Argsn: 1,
		Doc:   "Navigates forward to the next page.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					if forwarder, ok := bow.(interface{ Forward() error }); ok {
						err := forwarder.Forward()
						if err != nil {
							return evaldo.MakeError(ps, fmt.Sprintf("Failed to go forward: %s", err.Error()))
						}
						return *env.NewInteger(1)
					}
					return evaldo.MakeError(ps, "Browser does not support Forward method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// browser .reload
	// Args:
	// * browser: Surf browser instance
	// Returns:
	// * integer - 1 on success
	"surf-browser//Reload": {
		Argsn: 1,
		Doc:   "Reloads the current page.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					if reloader, ok := bow.(interface{ Reload() error }); ok {
						err := reloader.Reload()
						if err != nil {
							return evaldo.MakeError(ps, fmt.Sprintf("Failed to reload: %s", err.Error()))
						}
						return *env.NewInteger(1)
					}
					return evaldo.MakeError(ps, "Browser does not support Reload method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// url: browser .url
	// Args:
	// * browser: Surf browser instance
	// Returns:
	// * string - Current page URL
	"surf-browser//Url": {
		Argsn: 1,
		Doc:   "Gets the current URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					if urler, ok := bow.(interface{ Url() interface{} }); ok {
						url := urler.Url()
						if url != nil {
							if urlStringer, ok := url.(interface{ String() string }); ok {
								return *env.NewString(urlStringer.String())
							}
							return *env.NewString(fmt.Sprintf("%v", url))
						}
						return *env.NewString("")
					}
					return evaldo.MakeError(ps, "Browser does not support Url method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf browser.")
			}
		},
	},

	//
	// ##### Bookmarks ##### "Functions for managing bookmarks"
	//

	// Tests:
	// browser .bookmark "my-page"
	// Args:
	// * browser: Surf browser instance
	// * name: String - Name for the bookmark
	// Returns:
	// * integer - 1 on success
	"surf-browser//Bookmark": {
		Argsn: 2,
		Doc:   "Bookmarks the current page with a name.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch name := arg1.(type) {
					case env.String:
						if bookmarker, ok := bow.(interface{ Bookmark(string) }); ok {
							bookmarker.Bookmark(name.Value)
							return arg0
						}
						return evaldo.MakeError(ps, "Browser does not support Bookmark method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (bookmark name).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// browser .open-bookmark "my-page"
	// Args:
	// * browser: Surf browser instance
	// * name: String - Name of the bookmark to open
	// Returns:
	// * integer - 1 on success
	"surf-browser//Open-bookmark": {
		Argsn: 2,
		Doc:   "Opens a previously bookmarked page.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch name := arg1.(type) {
					case env.String:
						if bookmarkOpener, ok := bow.(interface{ OpenBookmark(string) error }); ok {
							err := bookmarkOpener.OpenBookmark(name.Value)
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Failed to open bookmark: %s", err.Error()))
							}
							return arg0
						}
						return evaldo.MakeError(ps, "Browser does not support OpenBookmark method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (bookmark name).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	//
	// ##### Forms and Element Interaction ##### "Functions for working with forms and page elements"
	//

	// Tests:
	// form: browser .form "#login-form"
	// Args:
	// * browser: Surf browser instance
	// * selector: String - CSS selector of the form
	// Returns:
	// * surf-form - Form object for further manipulation
	"surf-browser//Form": {
		Argsn: 2,
		Doc:   "Gets a form by CSS selector.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch selector := arg1.(type) {
					case env.String:
						if formGetter, ok := bow.(interface {
							Form(string) (interface{}, error)
						}); ok {
							form, err := formGetter.Form(selector.Value)
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Failed to get form: %s", err.Error()))
							}
							return *env.NewNative(ps.Idx, form, "surf-form")
						}
						return evaldo.MakeError(ps, "Browser does not support Form method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (CSS selector).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	// Tests:
	// form .input "username" "john_doe"
	// Args:
	// * form: Surf form instance
	// * name: String - Name of the input field
	// * value: String - Value to set in the input
	// Returns:
	// * integer - 1 on success
	"surf-form//Input": {
		Argsn: 3,
		Doc:   "Sets input value in a form.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch form := arg0.(type) {
			case env.Native:
				fm := form.Value
				if fm != nil {
					switch name := arg1.(type) {
					case env.String:
						switch value := arg2.(type) {
						case env.String:
							if inputter, ok := fm.(interface{ Input(string, string) error }); ok {
								err := inputter.Input(name.Value, value.Value)
								if err != nil {
									return evaldo.MakeError(ps, fmt.Sprintf("Failed to set input: %s", err.Error()))
								}
								return *env.NewInteger(1)
							}
							return evaldo.MakeError(ps, "Form does not support Input method.")
						default:
							return evaldo.MakeError(ps, "Third argument must be a string (value).")
						}
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (input name).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf form.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf form.")
			}
		},
	},

	// Tests:
	// form .submit
	// Args:
	// * form: Surf form instance
	// Returns:
	// * integer - 1 on success
	"surf-form//Submit": {
		Argsn: 1,
		Doc:   "Submits the form.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch form := arg0.(type) {
			case env.Native:
				fm := form.Value
				if fm != nil {
					if submitter, ok := fm.(interface{ Submit() error }); ok {
						err := submitter.Submit()
						if err != nil {
							return evaldo.MakeError(ps, fmt.Sprintf("Failed to submit form: %s", err.Error()))
						}
						return *env.NewInteger(1)
					}
					return evaldo.MakeError(ps, "Form does not support Submit method.")
				}
				return evaldo.MakeError(ps, "Argument must be a surf form.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf form.")
			}
		},
	},

	"surf-browser//Find": {
		Argsn: 2,
		Doc:   "Finds elements by CSS selector and returns a selection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch selector := arg1.(type) {
					case env.String:
						if finder, ok := bow.(interface {
							Find(string) *goquery.Selection
						}); ok {
							selection := finder.Find(selector.Value)
							return *env.NewNative(ps.Idx, selection, "surf-selection")
						}
						return evaldo.MakeError(ps, "Browser does not support Find method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (CSS selector).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	// TODO deduplice with above
	"surf-selection//Find": {
		Argsn: 2,
		Doc:   "Finds elements by CSS selector and returns a selection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch selector := arg1.(type) {
					case env.String:
						if finder, ok := bow.(interface {
							Find(string) *goquery.Selection
						}); ok {
							selection := finder.Find(selector.Value)
							return *env.NewNative(ps.Idx, selection, "surf-selection")
						}
						return evaldo.MakeError(ps, "Browser does not support Find method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (CSS selector).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	"surf-selection//Each": {
		Argsn: 2,
		Doc:   "Iterates over each element in the selection, calling the block with the goquery.Selection object for each element.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch selection := arg0.(type) {
			case env.Native:
				if sel, ok := selection.Value.(*goquery.Selection); ok {
					switch block := arg1.(type) {
					case env.Block:
						sel.Each(func(i int, s *goquery.Selection) {
							// Create a new environment and evaluate the block
							ser := ps.Ser
							ps.Ser = block.Series

							// Inject the goquery.Selection object (not text string)
							// This allows calling methods like .text?, .find, etc. on each element
							evaldo.EvalBlockInj(ps, *env.NewNative(ps.Idx, s, "surf-selection"), true)

							ps.Ser = ser
						})
						return *env.NewInteger(1)
					default:
						return evaldo.MakeError(ps, "Second argument must be a block (callback function).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf selection.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf selection.")
			}
		},
	},

	"surf-selection//Text?": {
		Argsn: 1,
		Doc:   "Gets the combined text of all elements in the selection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch selection := arg0.(type) {
			case env.Native:
				if sel, ok := selection.Value.(*goquery.Selection); ok {
					text := sel.Text()
					return *env.NewString(text)
				}
				return evaldo.MakeError(ps, "Argument must be a surf selection.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf selection.")
			}
		},
	},

	"surf-selection//Length?": {
		Argsn: 1,
		Doc:   "Gets the number of elements in the selection.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch selection := arg0.(type) {
			case env.Native:
				if sel, ok := selection.Value.(*goquery.Selection); ok {
					length := sel.Length()
					return *env.NewInteger(int64(length))
				}
				return evaldo.MakeError(ps, "Argument must be a surf selection.")
			default:
				return evaldo.MakeError(ps, "Argument must be a surf selection.")
			}
		},
	},

	//
	// ##### Configuration ##### "Functions for configuring browser settings"
	//

	// Tests:
	// browser .set-user-agent "Mozilla/5.0 (compatible; Bot/1.0)"
	// Args:
	// * browser: Surf browser instance
	// * user-agent: String - User agent string to set
	// Returns:
	// * integer - 1 on success
	"surf-browser//Set-user-agent": {
		Argsn: 2,
		Doc:   "Sets the user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch userAgent := arg1.(type) {
					case env.String:
						if agentSetter, ok := bow.(interface{ SetUserAgent(string) }); ok {
							agentSetter.SetUserAgent(userAgent.Value)
							return arg0
						}
						return evaldo.MakeError(ps, "Browser does not support SetUserAgent method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (user agent).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	//
	// ##### User Agents ##### "Predefined user agent strings for common browsers"
	//

	// Tests:
	// ua: surf-agent-chrome
	// Args:
	// Returns:
	// * string - Chrome user agent string
	"surf-agent-chrome": {
		Argsn: 0,
		Doc:   "Returns the Chrome user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (Windows NT 6.3; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36")
		},
	},

	// Tests:
	// ua: surf-agent-firefox
	// Args:
	// Returns:
	// * string - Firefox user agent string
	"surf-agent-firefox": {
		Argsn: 0,
		Doc:   "Returns the Firefox user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (Windows NT 6.3; x64; rv:31.0) Gecko/20100101 Firefox/31.0")
		},
	},

	// Tests:
	// ua: surf-agent-safari
	// Args:
	// Returns:
	// * string - Safari user agent string
	"surf-agent-safari": {
		Argsn: 0,
		Doc:   "Returns the Safari user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Safari/8536.25")
		},
	},

	// Tests:
	// ua: surf-agent-opera
	// Args:
	// Returns:
	// * string - Opera user agent string
	"surf-agent-opera": {
		Argsn: 0,
		Doc:   "Returns the Opera user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (Windows NT 6.3; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36 OPR/24.0.1558.64")
		},
	},

	// Tests:
	// ua: surf-agent-msie
	// Args:
	// Returns:
	// * string - MSIE user agent string
	"surf-agent-msie": {
		Argsn: 0,
		Doc:   "Returns the MSIE (Internet Explorer) user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)")
		},
	},

	// Tests:
	// ua: surf-agent-googlebot
	// Args:
	// Returns:
	// * string - GoogleBot user agent string
	"surf-agent-googlebot": {
		Argsn: 0,
		Doc:   "Returns the GoogleBot user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
		},
	},

	// Tests:
	// ua: surf-agent-bingbot
	// Args:
	// Returns:
	// * string - BingBot user agent string
	"surf-agent-bingbot": {
		Argsn: 0,
		Doc:   "Returns the BingBot user agent string.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewString("Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)")
		},
	},

	// Tests:
	// ua: surf-agent-create-version "chrome" "120"
	// Args:
	// * browser: String - Browser name (chrome, firefox, safari, etc.)
	// * version: String - Version number
	// Returns:
	// * string - Custom user agent string with specified version
	"surf-agent-create-version": {
		Argsn: 2,
		Doc:   "Creates a custom user agent string with a specific browser and version.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.String:
				switch version := arg1.(type) {
				case env.String:
					browserLower := strings.ToLower(browser.Value)
					switch browserLower {
					case "chrome":
						return *env.NewString(fmt.Sprintf("Mozilla/5.0 (Windows NT 6.3; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", version.Value))
					case "firefox":
						return *env.NewString(fmt.Sprintf("Mozilla/5.0 (Windows NT 6.3; x64; rv:%s) Gecko/20100101 Firefox/%s", version.Value, version.Value))
					case "safari":
						return *env.NewString(fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/536.26 (KHTML, like Gecko) Version/%s Safari/8536.25", version.Value))
					case "opera":
						return *env.NewString(fmt.Sprintf("Mozilla/5.0 (Windows NT 6.3; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36 OPR/%s", version.Value))
					default:
						return evaldo.MakeError(ps, fmt.Sprintf("Unknown browser: %s. Supported browsers: chrome, firefox, safari, opera", browser.Value))
					}
				default:
					return evaldo.MakeError(ps, "Second argument must be a string (version).")
				}
			default:
				return evaldo.MakeError(ps, "First argument must be a string (browser name).")
			}
		},
	},

	"surf-browser//Set-cookie": {
		Argsn: 2,
		Doc:   "Sets a cookie.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch cookie := arg1.(type) {
					case env.Dict:
						if cookieSetter, ok := bow.(interface{ SetCookie(interface{}) }); ok {
							// Convert Rye dict to http.Cookie
							httpCookie := &http.Cookie{}

							if name, exists := cookie.Data["name"]; exists {
								if nameStr, ok := name.(env.String); ok {
									httpCookie.Name = nameStr.Value
								}
							}

							if value, exists := cookie.Data["value"]; exists {
								if valueStr, ok := value.(env.String); ok {
									httpCookie.Value = valueStr.Value
								}
							}

							if domain, exists := cookie.Data["domain"]; exists {
								if domainStr, ok := domain.(env.String); ok {
									httpCookie.Domain = domainStr.Value
								}
							}

							if path, exists := cookie.Data["path"]; exists {
								if pathStr, ok := path.(env.String); ok {
									httpCookie.Path = pathStr.Value
								}
							}

							cookieSetter.SetCookie(httpCookie)
							return *env.NewInteger(1)
						}
						return evaldo.MakeError(ps, "Browser does not support SetCookie method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a dictionary (cookie).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},

	"surf-browser//Download": {
		Argsn: 2,
		Doc:   "Downloads a file from the given URL.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch browser := arg0.(type) {
			case env.Native:
				bow := browser.Value
				if bow != nil {
					switch url := arg1.(type) {
					case env.String:
						if downloader, ok := bow.(interface{ Download(string) error }); ok {
							err := downloader.Download(url.Value)
							if err != nil {
								return evaldo.MakeError(ps, fmt.Sprintf("Failed to download: %s", err.Error()))
							}
							return *env.NewInteger(1)
						}
						return evaldo.MakeError(ps, "Browser does not support Download method.")
					default:
						return evaldo.MakeError(ps, "Second argument must be a string (URL).")
					}
				}
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			default:
				return evaldo.MakeError(ps, "First argument must be a surf browser.")
			}
		},
	},
}
