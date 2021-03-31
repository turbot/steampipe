package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/steampipe/workspace"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/autocomplete"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/definitions/results"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/metaquery"
	"github.com/turbot/steampipe/queryhistory"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/version"

	"github.com/c-bata/go-prompt"
)

// InteractiveClient :: wrapper over *Client and *prompt.Prompt along
// to facilitate interactive query prompt
type InteractiveClient struct {
	client                  *Client
	workspace               *workspace.Workspace
	interactiveBuffer       []string
	interactivePrompt       *prompt.Prompt
	interactiveQueryHistory *queryhistory.QueryHistory
	autocompleteOnEmpty     bool
}

func newInteractiveClient(client *Client, workspace *workspace.Workspace) (*InteractiveClient, error) {
	return &InteractiveClient{
		client:                  client,
		workspace:               workspace,
		interactiveQueryHistory: queryhistory.New(),
		interactiveBuffer:       []string{},
		autocompleteOnEmpty:     false,
	}, nil
}

func (c *InteractiveClient) close() {
	// close the underlying client
	c.client.close()
}

// InteractiveQuery :: start an interactive prompt and return
func (c *InteractiveClient) InteractiveQuery(resultsStreamer *results.ResultStreamer, onCompleteCallback func()) {
	defer func() {
		onCompleteCallback()
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}

		// close the result stream
		// this needs to be the last thing we do,
		// as the runQueryCmd uses this as an indication
		// to quit out of the application
		resultsStreamer.Close()
	}()

	fmt.Printf("Welcome to Steampipe v%s\n", version.String())
	fmt.Printf("For more information, type %s\n", constants.Bold(".help"))

	for {
		rerun := c.runInteractivePrompt(resultsStreamer)

		// persist saved history
		c.interactiveQueryHistory.Persist()
		if !rerun.Restart {
			break
		}

		// wait for the resultstreamer to have streamed everything out
		// so that we know
		resultsStreamer.Wait()
	}
}

func (c *InteractiveClient) runInteractivePrompt(resultsStreamer *results.ResultStreamer) (ret utils.InteractiveExitStatus) {
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

	callExecutor := func(line string) { c.executor(line, resultsStreamer) }
	completer := func(d prompt.Document) []prompt.Suggest { return c.queryCompleter(d, c.client.schemaMetadata) }
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

func (c *InteractiveClient) executor(line string, resultsStreamer *results.ResultStreamer) {
	line = strings.TrimSpace(line)

	// if it's an empty line, then we don't need to do anything
	if line == "" {
		return
	}

	// push the current line into the buffer
	c.interactiveBuffer = append(c.interactiveBuffer, line)

	// expand the buffer out into 'query'
	query := strings.Join(c.interactiveBuffer, "\n")

	// if it is a multiline query, execute even without `;`
	if namedQuery, ok := c.workspace.GetNamedQuery(query); ok {
		query = namedQuery.SQL
	} else {
		// should we execute?
		if !c.shouldExecute(query) {
			return
		}
	}

	// so we need to execute - what are we executing

	// if the line is ONLY a semicolon, do nothing and restart interactive session
	if strings.TrimSpace(query) == ";" {
		resultsStreamer.Done()
		c.restartInteractiveSession()
	}

	if metaquery.IsMetaQuery(query) {
		if err := c.executeMetaquery(query); err != nil {
			utils.ShowError(err)
		}
		resultsStreamer.Done()
	} else {
		// otherwise execute query
		shouldShowCounter := cmdconfig.Viper().GetString(constants.ArgOutput) == constants.ArgTable
		if result, err := c.client.executeQuery(query, shouldShowCounter); err != nil {
			utils.ShowError(err)
			resultsStreamer.Done()
		} else {
			resultsStreamer.StreamResult(result)
		}
	}

	// store the history
	c.interactiveQueryHistory.Put(query)
	c.restartInteractiveSession()
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
		s = append(s, c.nameQuerySuggestions()...)
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

func (c *InteractiveClient) nameQuerySuggestions() []prompt.Suggest {
	var res []prompt.Suggest
	// add all the queries in the workspace
	for name, q := range c.workspace.GetNamedQueryMap() {
		description := "named query"
		if q.Description != "" {
			description += fmt.Sprintf(": %s", q.Description)
		}
		res = append(res, prompt.Suggest{Text: name, Description: description})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})
	return res
}
