package interactive

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sort"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/contexthelpers"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/query"
	"github.com/turbot/steampipe/query/metaquery"
	"github.com/turbot/steampipe/query/queryhistory"
	"github.com/turbot/steampipe/query/queryresult"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/statushooks"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"
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
	afterClose     AfterPromptCloseAction
	// lock while execution is occurring to avoid errors/warnings being shown
	executionLock  sync.Mutex
	schemaMetadata *schema.Metadata

	highlighter *Highlighter

	// status update hooks
	statusHook statushooks.StatusHooks
}

func getHighlighter(theme string) *Highlighter {
	return newHighlighter(
		lexers.Get("sql"),
		formatters.Get("terminal256"),
		styles.Native,
	)
}

func newInteractiveClient(ctx context.Context, initData *query.InitData, resultsStreamer *queryresult.ResultStreamer) (*InteractiveClient, error) {
	c := &InteractiveClient{
		initData:                initData,
		resultsStreamer:         resultsStreamer,
		interactiveQueryHistory: queryhistory.New(),
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
func (c *InteractiveClient) InteractivePrompt(ctx context.Context) {
	// start a cancel handler for the interactive client - this will call activeQueryCancelFunc if it is set
	// (registered when we call createQueryContext)
	interruptSignalChannel := contexthelpers.StartCancelHandler(c.cancelActiveQueryIfAny)

	// create a cancel context for the prompt - this will set c.cancelPrompt
	parentContext := ctx
	ctx = c.createPromptContext(parentContext)

	defer func() {
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
		}
		// close up the SIGINT channel so that the receiver goroutine can quit
		signal.Stop(interruptSignalChannel)
		close(interruptSignalChannel)

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
			// create new context
			ctx = c.createPromptContext(parentContext)
			// now run it again
			c.runInteractivePromptAsync(ctx, &promptResultChan)
		}
	}
}

// ClosePrompt cancels the running prompt, setting the action to take after close
func (c *InteractiveClient) ClosePrompt(afterClose AfterPromptCloseAction) {
	c.afterClose = afterClose
	c.cancelPrompt()
}

// LoadSchema implements Client
// retrieve both the raw query result and a sanitised version in list form
func (c *InteractiveClient) LoadSchema() error {
	utils.LogTime("db_client.LoadSchema start")
	defer utils.LogTime("db_client.LoadSchema end")

	// build a ConnectionSchemaMap object to identify the schemas to load
	// (pass nil for connection state - this forces NewConnectionSchemaMap to load it)
	connectionSchemaMap, err := steampipeconfig.NewConnectionSchemaMap()
	if err != nil {
		return err
	}
	// get the unique schema - we use this to limit the schemas we load from the database
	schemas := connectionSchemaMap.UniqueSchemas()
	// load these schemas
	// in a background context, since we are not running in a context - but GetSchemaFromDB needs one
	metadata, err := c.client().GetSchemaFromDB(context.Background(), schemas)
	if err != nil {
		return err
	}

	c.populateSchemaMetadata(metadata, connectionSchemaMap)

	return nil
}

// init data has arrived, handle any errors/warnings/messages
func (c *InteractiveClient) handleInitResult(ctx context.Context, initResult *db_common.InitResult) {
	// try to take an execution lock, so that we don't end up showing warnings and errors
	// while an execution is underway
	c.executionLock.Lock()
	defer c.executionLock.Unlock()

	if utils.IsContextCancelled(ctx) {
		log.Printf("[TRACE] prompt context has been cancelled - not handling init result")
		return
	}

	if initResult.Error != nil {
		c.ClosePrompt(AfterPromptCloseExit)
		// add newline to ensure error is not printed at end of current prompt line
		fmt.Println()
		utils.ShowError(ctx, initResult.Error)
		return
	}

	if initResult.HasMessages() {
		fmt.Println()
		initResult.DisplayMessages()
	}

	// We need to render the prompt here to make sure that it comes back
	// after the messages have been displayed
	c.interactivePrompt.Render()

	// tell the workspace to reset the prompt after displaying async filewatcher messages
	c.initData.Workspace.SetOnFileWatcherEventMessages(func() { c.interactivePrompt.Render() })
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
	// store the history (the raw line which was entered)
	// we want to store even if we fail to resolve a query
	c.interactiveQueryHistory.Push(line)

	query, err := c.getQuery(ctx, line)
	if query == "" {
		if err != nil {
			utils.ShowError(ctx, utils.HandleCancelError(err))
		}
		// restart the prompt
		c.restartInteractiveSession()
		return
	}

	// create a  context for the execution of the query
	queryContext := c.createQueryContext(ctx)

	if metaquery.IsMetaQuery(query) {
		if err := c.executeMetaquery(queryContext, query); err != nil {
			utils.ShowError(ctx, err)
		}
		// cancel the context
		c.cancelActiveQueryIfAny()

	} else {
		// otherwise execute query
		result, err := c.client().Execute(queryContext, query)
		if err != nil {
			utils.ShowError(ctx, utils.HandleCancelError(err))
		} else {
			c.resultsStreamer.StreamResult(result)
		}
	}

	// restart the prompt
	c.restartInteractiveSession()
}

func (c *InteractiveClient) getQuery(ctx context.Context, line string) (string, error) {
	// if it's an empty line, then we don't need to do anything
	if line == "" {
		return "", nil
	}

	// wait for initialisation to complete so we can access the workspace
	if !c.isInitialised() {
		// create a context used purely to detect cancellation during initialisation
		// this will also set c.cancelActiveQuery
		queryContext := c.createQueryContext(ctx)
		defer func() {
			// cancel this context
			c.cancelActiveQueryIfAny()
		}()

		statushooks.SetStatus(ctx, "Initializing...")
		// wait for client initialisation to complete
		err := c.waitForInitData(queryContext)
		statushooks.Done(ctx)
		if err != nil {
			// if it failed, report error and quit
			return "", err
		}
	}

	// push the current line into the buffer
	c.interactiveBuffer = append(c.interactiveBuffer, line)

	// expand the buffer out into 'query'
	queryString := strings.Join(c.interactiveBuffer, "\n")

	// in case of a named query call with params, parse the where clause
	query, _, err := c.workspace().ResolveQueryAndArgs(queryString)
	if err != nil {
		// if we fail to resolve, show error but do not return it - we want to stay in the prompt
		utils.ShowError(ctx, err)
		return "", nil
	}
	isNamedQuery := query != queryString

	// if it is a multiline query, execute even without `;`
	if !isNamedQuery {
		// should we execute?
		if !c.shouldExecute(queryString) {
			return "", nil
		}
	}

	// so we need to execute - what are we executing

	// if the line is ONLY a semicolon, do nothing and restart interactive session
	if strings.TrimSpace(query) == ";" {
		c.restartInteractiveSession()
		return "", nil
	}

	return query, nil
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
	// empty the buffer
	c.interactiveBuffer = []string{}
	// restart the prompt
	c.ClosePrompt(c.afterClose)
}

func (c *InteractiveClient) shouldExecute(line string) bool {
	return !cmdconfig.Viper().GetBool(constants.ArgMultiLine) || strings.HasSuffix(line, ";") || metaquery.IsMetaQuery(line)
}

func (c *InteractiveClient) queryCompleter(d prompt.Document) []prompt.Suggest {
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
		s = append(s, c.namedQuerySuggestions()...)
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

func (c *InteractiveClient) namedQuerySuggestions() []prompt.Suggest {
	var res []prompt.Suggest
	// only add named query suggestions if the client is initialised
	if !c.isInitialised() {
		return nil
	}
	resourceMaps := c.workspace().GetResourceMaps()
	// add all the queries in the workspace
	for queryName, q := range resourceMaps.LocalQueries {
		res = append(res, c.addQuerySuggestion(q, queryName))
	}
	for queryName, q := range resourceMaps.Queries {
		res = append(res, c.addQuerySuggestion(q, queryName))
	}

	// add all the controls in the workspace
	for controlName, control := range resourceMaps.LocalControls {
		res = append(res, c.addControlSuggestion(control, controlName))
	}
	for controlName, control := range resourceMaps.Controls {
		res = append(res, c.addControlSuggestion(control, controlName))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	return res
}

func (c *InteractiveClient) addQuerySuggestion(query *modconfig.Query, queryName string) prompt.Suggest {
	description := "named query"
	if query.Description != nil {
		description += fmt.Sprintf(": %s", *query.Description)
	}
	return prompt.Suggest{Text: queryName, Output: queryName, Description: description}
}

func (c *InteractiveClient) addControlSuggestion(control *modconfig.Control, controlName string) prompt.Suggest {
	description := "control"
	if control.Description != nil {
		description += fmt.Sprintf(": %s", *control.Description)
	}
	return prompt.Suggest{Text: controlName, Output: controlName, Description: description}
}

func (c *InteractiveClient) populateSchemaMetadata(schemaMetadata *schema.Metadata, connectionSchemaMap steampipeconfig.ConnectionSchemaMap) error {
	// we now need to add in all other schemas which have the same schemas as those we have loaded
	for loadedSchema, otherSchemas := range connectionSchemaMap {
		// all 'otherSchema's have the same schema as loadedSchema
		exemplarSchema, ok := schemaMetadata.Schemas[loadedSchema]
		if !ok {
			// should can happen in the case of a dynamic plugin with no tables - use empty schema
			exemplarSchema = make(map[string]schema.TableSchema)
		}

		for _, s := range otherSchemas {
			schemaMetadata.Schemas[s] = exemplarSchema
		}
	}
	c.schemaMetadata = schemaMetadata
	return nil
}
