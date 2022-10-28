package interactive

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/query"
	"github.com/turbot/steampipe/pkg/query/metaquery"
	"github.com/turbot/steampipe/pkg/query/queryhistory"
	"github.com/turbot/steampipe/pkg/query/queryresult"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/version"
)

type AfterPromptCloseAction int

const (
	AfterPromptCloseExit AfterPromptCloseAction = iota
	AfterPromptCloseRestart
)

// InteractiveClient is a wrapper over a LocalClient and a Prompt to facilitate interactive query prompt
type InteractiveClient struct {
	initData                *query.InitData
	resultsStreamer         *queryresult.ResultStreamer
	interactiveBuffer       []string
	interactivePrompt       *prompt.Prompt
	interactiveQueryHistory *queryhistory.QueryHistory
	autocompleteOnEmpty     bool
	// the cancellation function for the active query - may be nil
	// NOTE: should ONLY be called by cancelActiveQueryIfAny
	cancelActiveQuery context.CancelFunc
	cancelPrompt      context.CancelFunc
	// channel used internally to pass the initialisation result
	initResultChan chan *db_common.InitResult
	// flag set when initialisation is complete (with or without errors)
	initialisationComplete bool
	afterClose             AfterPromptCloseAction
	// lock while execution is occurring to avoid errors/warnings being shown
	executionLock sync.Mutex
	// the schema metadata - this is loaded asynchronously during init
	schemaMetadata *schema.Metadata
	highlighter    *Highlighter
	hidePrompt     bool
	suggestions    []prompt.Suggest
}

func getHighlighter(theme string) *Highlighter {
	return newHighlighter(
		lexers.Get("sql"),
		formatters.Get("terminal256"),
		styles.Native,
	)
}

func newInteractiveClient(ctx context.Context, initData *query.InitData, resultsStreamer *queryresult.ResultStreamer) (*InteractiveClient, error) {
	interactiveQueryHistory, err := queryhistory.New()
	if err != nil {
		return nil, err
	}
	c := &InteractiveClient{
		initData:                initData,
		resultsStreamer:         resultsStreamer,
		interactiveQueryHistory: interactiveQueryHistory,
		interactiveBuffer:       []string{},
		autocompleteOnEmpty:     false,
		initResultChan:          make(chan *db_common.InitResult, 1),
		highlighter:             getHighlighter(viper.GetString(constants.ArgTheme)),
	}

	// asynchronously wait for init to complete
	// we start this immediately rather than lazy loading as we want to handle errors asap
	go c.readInitDataStream(ctx)
	return c, nil
}

// InteractivePrompt starts an interactive prompt and return
func (c *InteractiveClient) InteractivePrompt(parentContext context.Context) {
	// start a cancel handler for the interactive client - this will call activeQueryCancelFunc if it is set
	// (registered when we call createQueryContext)
	quitChannel := c.startCancelHandler()

	// create a cancel context for the prompt - this will set c.cancelPrompt
	ctx := c.createPromptContext(parentContext)

	defer func() {
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
		// close up the SIGINT channel so that the receiver goroutine can quit
		quitChannel <- true
		close(quitChannel)

		// cleanup the init data to ensure any services we started are stopped
		c.initData.Cleanup(ctx)

		// close the result stream
		// this needs to be the last thing we do,
		// as the query result display code will exit once the result stream is closed
		c.resultsStreamer.Close()
	}()

	statushooks.Message(ctx,
		fmt.Sprintf("Welcome to Steampipe v%s", version.SteampipeVersion.String()),
		fmt.Sprintf("For more information, type %s", constants.Bold(".help")))

	// run the prompt in a goroutine, so we can also detect async initialisation errors
	promptResultChan := make(chan utils.InteractiveExitStatus, 1)
	c.runInteractivePromptAsync(ctx, &promptResultChan)

	// select results
	for {
		select {
		case initResult := <-c.initResultChan:
			c.handleInitResult(ctx, initResult)
			// if there was an error, handleInitResult will shut down the prompt
			// - we must wait for it to shut down and not return immediately

		case <-promptResultChan:
			// persist saved history
			c.interactiveQueryHistory.Persist()
			// check post-close action
			if c.afterClose == AfterPromptCloseExit {
				return
			}
			// create new context with a cancellation func
			ctx = c.createPromptContext(parentContext)
			// now run it again
			c.runInteractivePromptAsync(ctx, &promptResultChan)
		}
	}
}

func (c *InteractiveClient) initialiseSuggestions() {
	var res []prompt.Suggest

	workspaceModName := c.initData.Workspace.Mod.Name()
	resourceFunc := func(item modconfig.HclResource) (continueWalking bool, err error) {
		continueWalking = true

		qp, ok := item.(modconfig.QueryProvider)
		if !ok {
			return
		}
		if qp.GetQuery() == nil && qp.GetSQL() == nil {
			return
		}
		rm := item.(modconfig.ResourceWithMetadata)
		if rm.IsAnonymous() {
			return
		}
		isLocal := qp.GetMod().Name() == workspaceModName
		itemType := item.BlockType()
		// only include global inputs
		if itemType == modconfig.BlockTypeInput {
			if _, ok := c.initData.Workspace.Mod.ResourceMaps.GlobalDashboardInputs[item.Name()]; !ok {
				return
			}
		}
		// special case for query
		if itemType == modconfig.BlockTypeQuery {
			itemType = "named query"
		}
		name := qp.Name()
		if isLocal {
			name = qp.GetUnqualifiedName()
		}

		res = append(res, c.addSuggestion(itemType, qp.GetDescription(), name))
		return
	}

	c.workspace().GetResourceMaps().WalkResources(resourceFunc)

	// sort the suggestions
	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	c.suggestions = res
}

// ClosePrompt cancels the running prompt, setting the action to take after close
func (c *InteractiveClient) ClosePrompt(afterClose AfterPromptCloseAction) {
	c.afterClose = afterClose
	c.cancelPrompt()
}

// retrieve both the raw query result and a sanitised version in list form
func (c *InteractiveClient) loadSchema() error {
	utils.LogTime("db_client.loadSchema start")
	defer utils.LogTime("db_client.loadSchema end")

	// load these schemas
	// in a background context, since we are not running in a context - but GetSchemaFromDB needs one
	metadata, err := c.client().GetSchemaFromDB(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load schemas: %s", err.Error())
	}

	c.schemaMetadata = metadata
	return nil
}

func (c *InteractiveClient) runInteractivePromptAsync(ctx context.Context, promptResultChan *chan utils.InteractiveExitStatus) {
	go func() {
		*promptResultChan <- c.runInteractivePrompt(ctx)
	}()
}

func (c *InteractiveClient) runInteractivePrompt(ctx context.Context) (ret utils.InteractiveExitStatus) {
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
			if r != nil {
				// for everything else, float up the panic
				panic(r)
			}
		}
	}()

	callExecutor := func(line string) {
		c.executor(ctx, line)
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return c.queryCompleter(d)
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
			if c.hidePrompt {
				prefix = ""
			}
			return
		}),
		prompt.OptionFormatter(c.highlighter.Highlight),
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
			Key: prompt.ControlD,
			Fn: func(b *prompt.Buffer) {
				if b.Text() == "" {
					// just set after close action - go prompt will handle the prompt shutdown
					c.afterClose = AfterPromptCloseExit
				}
			},
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
		prompt.OptionBufferPreHook(func(s string) (string, bool) {
			isWsl, err := utils.IsWSL()
			// if we got an error or this is not WSL, return as-is
			if err != nil || !isWsl {
				return s, false
			}
			b := []byte(s)
			// in WSL, 'Alt' combo-characters are denoted by [27, ASCII of character]
			// if we get a combination which has 27 as prefix - we should ignore it
			// this is inline with other interactive clients like pgcli
			if len(b) > 1 && bytes.HasPrefix(b, []byte{byte(27)}) {
				// ignore it
				return "", true
			}
			return string(b), false
		}),
	)
	// set this to a default
	c.autocompleteOnEmpty = false
	c.interactivePrompt.RunCtx(ctx)

	return
}

func (c *InteractiveClient) breakMultilinePrompt(buffer *prompt.Buffer) {
	c.interactiveBuffer = []string{}
}

func (c *InteractiveClient) executor(ctx context.Context, line string) {
	// take an execution lock, so that errors and warnings don't show up while
	// we are underway
	c.executionLock.Lock()
	defer c.executionLock.Unlock()

	// set afterClose to restart - is we are exiting the metaquery will set this to AfterPromptCloseExit
	c.afterClose = AfterPromptCloseRestart

	line = strings.TrimSpace(line)

	query := c.getQuery(ctx, line)
	if query == "" {
		// we failed to resolve a query, or are in the middle of a multi-line entry
		// restart the prompt, DO NOT clear the interactive buffer
		c.restartInteractiveSession()
		return
	}

	// we successfully retrieved a query

	// create a  context for the execution of the query
	queryCtx := c.createQueryContext(ctx)

	if metaquery.IsMetaQuery(query) {
		if err := c.executeMetaquery(queryCtx, query); err != nil {
			error_helpers.ShowError(ctx, err)
		}
		// cancel the context
		c.cancelActiveQueryIfAny()

	} else {
		// otherwise execute query
		t := time.Now()
		result, err := c.client().Execute(queryCtx, query)
		if err != nil {
			error_helpers.ShowError(ctx, error_helpers.HandleCancelError(err))
			// if timing flag is enabled, show the time taken for the query to fail
			if cmdconfig.Viper().GetBool(constants.ArgTiming) {
				display.DisplayErrorTiming(t)
			}
		} else {
			c.resultsStreamer.StreamResult(result)
		}
	}

	// restart the prompt
	c.restartInteractiveSession()
}

func (c *InteractiveClient) getQuery(ctx context.Context, line string) string {
	// if it's an empty line, then we don't need to do anything
	if line == "" {
		return ""
	}

	// store the history (the raw line which was entered)
	historyEntry := line
	defer func() {
		if len(historyEntry) > 0 {
			// we want to store even if we fail to resolve a query
			c.interactiveQueryHistory.Push(historyEntry)
		}

	}()

	// wait for initialisation to complete so we can access the workspace
	if !c.isInitialised() {
		// create a context used purely to detect cancellation during initialisation
		// this will also set c.cancelActiveQuery
		queryCtx := c.createQueryContext(ctx)
		defer func() {
			// cancel this context
			c.cancelActiveQueryIfAny()
		}()

		statushooks.SetStatus(ctx, "Initializing...")
		// wait for client initialisation to complete
		err := c.waitForInitData(queryCtx)
		statushooks.Done(ctx)
		if err != nil {
			// clear history entry
			historyEntry = ""
			// clear the interactive buffer
			c.interactiveBuffer = nil
			// error will have been handled elsewhere
			return ""
		}
	}

	// push the current line into the buffer
	c.interactiveBuffer = append(c.interactiveBuffer, line)

	// expand the buffer out into 'query'
	queryString := strings.Join(c.interactiveBuffer, "\n")

	// in case of a named query call with params, parse the where clause
	query, queryProvider, err := c.workspace().ResolveQueryAndArgsFromSQLString(queryString)
	if err != nil {
		// if we fail to resolve:
		// - show error but do not return it so we  stay in the prompt
		// - do not clear history item - we want to store bad entry in history
		// - clear interactive buffer
		c.interactiveBuffer = nil
		error_helpers.ShowError(ctx, err)
		return ""
	}
	isNamedQuery := queryProvider != nil

	// should we execute?
	// we will NOT execute if we are in multiline mode, there is no semi-colon
	// and it is NOT a metaquery or a named query
	if !c.shouldExecute(queryString, isNamedQuery) {
		// is we are not executing, do not store history
		historyEntry = ""
		// do not clear interactive buffer
		return ""
	}

	// so we need to execute
	// clear the interactive buffer
	c.interactiveBuffer = nil

	// what are we executing?

	// if the line is ONLY a semicolon, do nothing and restart interactive session
	if strings.TrimSpace(query) == ";" {
		// do not store in history
		historyEntry = ""
		c.restartInteractiveSession()
		return ""
	}
	// if this is a multiline query, update history entry
	if !isNamedQuery && len(strings.Split(query, "\n")) > 1 {
		historyEntry = query
	}

	return query
}

func (c *InteractiveClient) executeMetaquery(ctx context.Context, query string) error {
	// the client must be initialised to get here
	if !c.isInitialised() {
		panic("client is not initalised")
	}
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
	client := c.client()
	// validation passed, now we will run
	return metaquery.Handle(ctx, &metaquery.HandlerInput{
		Query:       query,
		Executor:    client,
		Schema:      c.schemaMetadata,
		Connections: client.ConnectionMap(),
		Prompt:      c.interactivePrompt,
		ClosePrompt: func() { c.afterClose = AfterPromptCloseExit },
	})
}

func (c *InteractiveClient) restartInteractiveSession() {
	// restart the prompt
	c.ClosePrompt(c.afterClose)
}

func (c *InteractiveClient) shouldExecute(line string, namedQuery bool) bool {
	if namedQuery {
		// execute named queries with no ';' even in multiline mode
		return true
	}
	if !cmdconfig.Viper().GetBool(constants.ArgMultiLine) {
		// NOT multiline mode
		return true
	}
	if metaquery.IsMetaQuery(line) {
		// execute metaqueries with no ';' even in multiline mode
		return true
	}
	if strings.HasSuffix(line, ";") {
		// statement has terminating ';'
		return true
	}

	return false
}

func (c *InteractiveClient) queryCompleter(d prompt.Document) []prompt.Suggest {
	if !cmdconfig.Viper().GetBool(constants.ArgAutoComplete) {
		return nil
	}
	if !c.isInitialised() {
		return nil
	}

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

		// named queries
		s = append(s, c.suggestions...)
		// "select"
		s = append(s, prompt.Suggest{Text: "select", Output: "select"}, prompt.Suggest{Text: "with", Output: "with"})

		// metaqueries
		s = append(s, metaquery.PromptSuggestions()...)

	} else if metaquery.IsMetaQuery(text) {
		client := c.client()
		suggestions := metaquery.Complete(&metaquery.CompleterInput{
			Query:            text,
			TableSuggestions: GetTableAutoCompleteSuggestions(c.schemaMetadata, client.ConnectionMap()),
		})

		s = append(s, suggestions...)
	} else {
		queryInfo := getQueryInfo(text)

		// only add table suggestions if the client is initialised
		if queryInfo.EditingTable && c.isInitialised() && c.schemaMetadata != nil {
			s = append(s, GetTableAutoCompleteSuggestions(c.schemaMetadata, c.initData.Client.ConnectionMap())...)
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

func (c *InteractiveClient) addSuggestion(itemType string, description string, name string) prompt.Suggest {
	if description != "" {
		itemType += fmt.Sprintf(": %s", description)
	}
	return prompt.Suggest{Text: name, Output: name, Description: itemType}
}

func (c *InteractiveClient) startCancelHandler() chan bool {
	sigIntChannel := make(chan os.Signal, 1)
	quitChannel := make(chan bool, 1)
	signal.Notify(sigIntChannel, os.Interrupt)
	go func() {
		for {
			select {
			case <-sigIntChannel:
				log.Println("[TRACE] interactive client cancel handler got SIGINT")
				// if initialisation is not complete, just close the prompt
				// this will cancel the context used for initialisation so cancel any initialisation queries
				if !c.isInitialised() {
					c.ClosePrompt(AfterPromptCloseExit)
					return
				} else {
					// otherwise call cancelActiveQueryIfAny which the for the active query, if there is one
					c.cancelActiveQueryIfAny()
					// keep waiting for further cancellations
				}
			case <-quitChannel:
				log.Println("[TRACE] cancel handler exiting")
				c.cancelActiveQueryIfAny()
				// we're done
				return
			}
		}
	}()
	return quitChannel
}
