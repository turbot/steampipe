package db

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/autocomplete"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/query/metaquery"
	"github.com/turbot/steampipe/query/queryhistory"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
)

// InteractiveClient :: wrapper over *Client and *prompt.Prompt along
// to facilitate interactive query prompt
type InteractiveClient struct {
	client                  *Client
	resultsStreamer         *queryresult.ResultStreamer
	workspace               WorkspaceResourceProvider
	interactiveBuffer       []string
	interactivePrompt       *prompt.Prompt
	interactiveQueryHistory *queryhistory.QueryHistory
	autocompleteOnEmpty     bool
	activeQueryCancelFunc   context.CancelFunc
}

func newInteractiveClient(workspace WorkspaceResourceProvider, client *Client, resultsStreamer *queryresult.ResultStreamer) (*InteractiveClient, error) {
	return &InteractiveClient{
		client:                  client,
		resultsStreamer:         resultsStreamer,
		workspace:               workspace,
		interactiveQueryHistory: queryhistory.New(),
		interactiveBuffer:       []string{},
		autocompleteOnEmpty:     false,
	}, nil
}

// InteractiveQuery :: start an interactive prompt and return
func (c *InteractiveClient) InteractiveQuery() {
	defer func() {
		// close the underlying client
		c.client.Close()
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}

		// close the result stream
		// this needs to be the last thing we do,
		// as the runQueryCmd uses this as an indication
		// to quit out of the application
		c.resultsStreamer.Close()
	}()

	interruptSignalChannel := c.startCancelHandler()
	fmt.Printf("Welcome to Steampipe v%s\n", version.String())
	fmt.Printf("For more information, type %s\n", constants.Bold(".help"))
	for {
		rerun := c.runInteractivePrompt()

		// persist saved history
		c.interactiveQueryHistory.Persist()
		if !rerun.Restart {
			break
		}

		// wait for the resultsStreamer to have streamed everything out
		// this is to be sure the previous command has completed streaming
		c.resultsStreamer.Wait()
	}

	// close up the SIGINT channel so that the receiver goroutine can quit
	close(interruptSignalChannel)
}

func (c *InteractiveClient) startCancelHandler() chan os.Signal {
	interruptSignalChannel := make(chan os.Signal, 10)
	signal.Notify(interruptSignalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range interruptSignalChannel {
			if c.hasActiveCancel() {
				c.activeQueryCancelFunc()
				c.clearCancelFunction()
			}
		}
	}()
	return interruptSignalChannel
}

func (c *InteractiveClient) runInteractivePrompt() (ret utils.InteractiveExitStatus) {
	defer func() {
		// this is to catch the PANIC that gets raised by
		// the executor of go-prompt
		//
		// We need to do it this way, since there is no
		// clean way to reload go-prompt so that we can
		// populate the history stack
		//
		r := recover()
		switch v := r.(type) {
		case utils.InteractiveExitStatus:
			// this is a planned exit
			// set the return value
			ret = v
		default:
			// for everything else, float up the panic
			panic(r)
		}
	}()

	callExecutor := func(line string) {
		c.executor(line)
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return c.queryCompleter(d, c.client.schemaMetadata)
	}
	c.interactivePrompt = prompt.New(
		callExecutor,
		completer,
		prompt.OptionTitle("steampipe interactive client "),
		prompt.OptionLivePrefix(func() (prefix string, useLive bool) {
			prefix = "> "
			useLive = true
			if len(c.interactiveBuffer) > 0 {
				prefix = ">>  "
			}
			return
		}),
		prompt.OptionHistory(c.interactiveQueryHistory.Get()),
		prompt.OptionInputTextColor(prompt.DefaultColor),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionMaxSuggestion(20),
		// Known Key Bindings
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn:  func(b *prompt.Buffer) { c.breakMultilinePrompt(b) },
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.Tab,
			Fn: func(b *prompt.Buffer) {
				if len(b.Text()) == 0 {
					c.autocompleteOnEmpty = true
				} else {
					c.autocompleteOnEmpty = false
				}
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.Escape,
			Fn: func(b *prompt.Buffer) {
				if len(b.Text()) == 0 {
					c.autocompleteOnEmpty = false
				}
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ShiftLeft,
			Fn:  prompt.GoLeftChar,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ShiftRight,
			Fn:  prompt.GoRightChar,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ShiftUp,
			Fn:  func(b *prompt.Buffer) { /*ignore*/ },
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ShiftDown,
			Fn:  func(b *prompt.Buffer) { /*ignore*/ },
		}),
		// Opt+LeftArrow
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: constants.OptLeftArrowASCIICode,
			Fn:        prompt.GoLeftWord,
		}),
		// Opt+RightArrow
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: constants.OptRightArrowASCIICode,
			Fn:        prompt.GoRightWord,
		}),
		// Alt+LeftArrow
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: constants.AltLeftArrowASCIICode,
			Fn:        prompt.GoLeftWord,
		}),
		// Alt+RightArrow
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: constants.AltRightArrowASCIICode,
			Fn:        prompt.GoRightWord,
		}),
	)
	// set this to a default
	c.autocompleteOnEmpty = false
	c.interactivePrompt.Run()
	return
}

func (c *InteractiveClient) breakMultilinePrompt(buffer *prompt.Buffer) {
	c.interactiveBuffer = []string{}
}

func (c *InteractiveClient) executor(line string) {
	line = strings.TrimSpace(line)

	// if it's an empty line, then we don't need to do anything
	if line == "" {
		return
	}
	// store history item before doing named query translation
	historyItem := line

	// push the current line into the buffer
	c.interactiveBuffer = append(c.interactiveBuffer, line)

	// expand the buffer out into 'query'
	query := strings.Join(c.interactiveBuffer, "\n")

	namedQuery, isNamedQuery := c.workspace.GetQuery(query)

	// if it is a multiline query, execute even without `;`
	if isNamedQuery {
		query = *namedQuery.SQL
	} else {
		// should we execute?
		if !c.shouldExecute(query) {
			return
		}
	}

	control, isControl := c.workspace.GetControl(query)
	if isControl {
		query = *control.SQL
	} else {
		// should we execute?
		if !c.shouldExecute(query) {
			return
		}
	}

	// so we need to execute - what are we executing

	// if the line is ONLY a semicolon, do nothing and restart interactive session
	if strings.TrimSpace(query) == ";" {
		c.restartInteractiveSession()
	}

	if metaquery.IsMetaQuery(query) {
		if err := c.executeMetaquery(query); err != nil {
			utils.ShowError(err)
		}
		c.resultsStreamer.Done()
	} else {
		// otherwise execute query
		ctx, cancel := context.WithCancel(context.Background())
		c.setCancelFunction(cancel)

		result, err := c.client.ExecuteQuery(ctx, query, false)
		if err != nil {
			c.handleExecuteError(err)
		} else {
			c.resultsStreamer.StreamResult(result)
		}
	}

	// store the history
	c.interactiveQueryHistory.Put(historyItem)
	c.restartInteractiveSession()
}

func (c *InteractiveClient) handleExecuteError(err error) {
	isCancelledError, isCancelledBeforeResult := isCancelledError(err)
	if isCancelledError {
		utils.ShowError(fmt.Errorf("execution cancelled"))
		if isCancelledBeforeResult {
			// we need to notify the streamer that we are done
			c.resultsStreamer.Done()
		}
	} else {
		utils.ShowError(err)
		c.resultsStreamer.Done()
	}
}

func isCancelledError(err error) (bool, bool) {

	isCancelledBeforeResult := strings.Contains(err.Error(), "Unrecognized remote plugin message")
	isCancelledUponResult := strings.Contains(err.Error(), "canceling statement due to user request")
	isCancelledAfterResult := err == context.Canceled

	isCancelledError := isCancelledBeforeResult || isCancelledUponResult || isCancelledAfterResult

	return isCancelledError, isCancelledBeforeResult
}

func (c *InteractiveClient) hasActiveCancel() bool {
	return c.activeQueryCancelFunc != nil
}

func (c *InteractiveClient) setCancelFunction(cancel context.CancelFunc) {
	c.activeQueryCancelFunc = cancel
}

func (c *InteractiveClient) clearCancelFunction() {
	c.activeQueryCancelFunc = nil
}

func (c *InteractiveClient) executeMetaquery(query string) error {
	// validate the metaquery arguments
	validateResult := metaquery.Validate(query)
	if validateResult.Message != "" {
		fmt.Println(validateResult.Message)
	}
	if err := validateResult.Err; err != nil {
		return err
	}
	if !validateResult.ShouldRun {
		return nil
	}
	// validation passed, now we will run
	return metaquery.Handle(&metaquery.HandlerInput{
		Query:       query,
		Executor:    c.client,
		Schema:      c.client.schemaMetadata,
		Connections: c.client.connectionMap,
		Prompt:      c.interactivePrompt,
	})
}

func (c *InteractiveClient) restartInteractiveSession() {
	// empty the buffer
	c.interactiveBuffer = []string{}
	// restart the prompt
	panic(utils.InteractiveExitStatus{Restart: true})
}

func (c *InteractiveClient) shouldExecute(line string) bool {
	return !cmdconfig.Viper().GetBool(constants.ArgMultiLine) || strings.HasSuffix(line, ";") || metaquery.IsMetaQuery(line)
}

func (c *InteractiveClient) queryCompleter(d prompt.Document, schemaMetadata *schema.Metadata) []prompt.Suggest {
	text := strings.TrimLeft(strings.ToLower(d.Text), " ")

	if len(c.interactiveBuffer) > 0 {
		text = strings.Join(append(c.interactiveBuffer, text), " ")
	}

	var s []prompt.Suggest

	if len(d.CurrentLine()) == 0 && !c.autocompleteOnEmpty {
		// if nothing has been typed yet, no point
		// giving suggestions
		return s
	}

	if isFirstWord(text) {
		// add all we know that can be the first words

		//named queries
		s = append(s, c.namedQuerySuggestions()...)
		// "select"
		s = append(s, prompt.Suggest{Text: "select"})

		// metaqueries
		s = append(s, metaquery.PromptSuggestions()...)

	} else if metaquery.IsMetaQuery(text) {
		suggestions := metaquery.Complete(&metaquery.CompleterInput{
			Query:       text,
			Schema:      c.client.schemaMetadata,
			Connections: c.client.connectionMap,
		})

		s = append(s, suggestions...)
	} else {
		queryInfo := getQueryInfo(text)

		if queryInfo.EditingTable {
			s = append(s, autocomplete.GetTableAutoCompleteSuggestions(c.client.schemaMetadata, c.client.connectionMap)...)
		}

		// Not sure this is working. comment out for now!
		// if queryInfo.EditingColumn {
		// 	fmt.Println(queryInfo.Table)
		// 	for _, column := range schemaMetadata.ColumnMap[queryInfo.Table] {
		// 		s = append(s, prompt.Suggest{Text: column})
		// 	}
		// }

	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func (c *InteractiveClient) namedQuerySuggestions() []prompt.Suggest {
	var res []prompt.Suggest
	// add all the queries in the workspace
	for queryName, q := range c.workspace.GetQueryMap() {
		description := "named query"
		if q.Description != nil {
			description += fmt.Sprintf(": %s", *q.Description)
		}
		res = append(res, prompt.Suggest{Text: queryName, Description: description})
	}
	// add all the controls in the workspace
	for controlName, c := range c.workspace.GetControlMap() {
		description := "control"
		if c.Description != nil {
			description += fmt.Sprintf(": %s", *c.Description)
		}
		res = append(res, prompt.Suggest{Text: controlName, Description: description})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	return res
}
