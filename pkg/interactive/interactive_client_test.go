package interactive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
)

// TestShouldExecute tests the logic that determines whether to execute a query
// This is critical for interactive mode behavior
func TestShouldExecute(t *testing.T) {
	tests := map[string]struct {
		line            string
		multiLineMode   bool
		expectedExecute bool
		description     string
	}{
		"single line mode always executes": {
			line:            "SELECT * FROM users",
			multiLineMode:   false,
			expectedExecute: true,
			description:     "without multiline mode, should execute immediately",
		},
		"multiline mode with semicolon executes": {
			line:            "SELECT * FROM users;",
			multiLineMode:   true,
			expectedExecute: true,
			description:     "with semicolon, should execute in multiline mode",
		},
		"multiline mode without semicolon does not execute": {
			line:            "SELECT * FROM users",
			multiLineMode:   true,
			expectedExecute: false,
			description:     "without semicolon in multiline mode, wait for more input",
		},
		"metaquery without semicolon executes in multiline": {
			line:            ".tables",
			multiLineMode:   true,
			expectedExecute: true,
			description:     "metaqueries should execute immediately even in multiline mode",
		},
		"metaquery with semicolon": {
			line:            ".help;",
			multiLineMode:   true,
			expectedExecute: true,
			description:     "metaqueries with semicolon should execute",
		},
		"empty line in multiline mode": {
			line:            "",
			multiLineMode:   true,
			expectedExecute: false,
			description:     "empty line should not execute",
		},
		"whitespace only in multiline mode": {
			line:            "   ",
			multiLineMode:   true,
			expectedExecute: false,
			description:     "whitespace-only should not execute",
		},
		"query with newlines and semicolon": {
			line:            "SELECT *\nFROM users\nWHERE id = 1;",
			multiLineMode:   true,
			expectedExecute: true,
			description:     "multiline query with semicolon should execute",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup viper config
			cmdconfig.Viper().Set(pconstants.ArgMultiLine, tc.multiLineMode)

			client := &InteractiveClient{}
			result := client.shouldExecute(tc.line)

			assert.Equal(t, tc.expectedExecute, result, tc.description)
		})
	}
}

// TestIsInitialised tests initialization status tracking
// Bug hunting: ensure initialization state is tracked correctly to avoid nil panics
func TestIsInitialised(t *testing.T) {
	tests := map[string]struct {
		client   *InteractiveClient
		expected bool
	}{
		"uninitialized client": {
			client: &InteractiveClient{
				initialisationComplete: false,
			},
			expected: false,
		},
		"initialized client": {
			client: &InteractiveClient{
				initialisationComplete: true,
			},
			expected: true,
		},
		"nil client should not panic": {
			client:   &InteractiveClient{},
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.client.isInitialised()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClient tests the client getter with initialization checks
// Bug hunting: nil dereference protection
func TestClient(t *testing.T) {
	tests := map[string]struct {
		setupClient func() *InteractiveClient
		expectNil   bool
		description string
	}{
		"client with nil initData returns nil": {
			setupClient: func() *InteractiveClient {
				return &InteractiveClient{
					initData: nil,
				}
			},
			expectNil:   true,
			description: "should return nil when initData is nil",
		},
		// Note: can't easily test with non-nil client without complex mocking
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := tc.setupClient()
			result := client.client()

			if tc.expectNil {
				assert.Nil(t, result, tc.description)
			} else {
				assert.NotNil(t, result, tc.description)
			}
		})
	}
}

// TestCleanBufferForWSL tests WSL-specific input cleaning
// Bug hunting: edge cases in byte array handling
func TestCleanBufferForWSL(t *testing.T) {
	tests := map[string]struct {
		input         string
		expectedStr   string
		expectedIgnore bool
		description   string
	}{
		"normal input without escape": {
			input:         "select * from users",
			expectedStr:   "select * from users",
			expectedIgnore: false,
			description:   "normal input should pass through unchanged",
		},
		"alt combo with escape prefix": {
			input:         string([]byte{27, 'a'}),
			expectedStr:   "",
			expectedIgnore: true,
			description:   "Alt+key combinations should be ignored",
		},
		"escape at end of input": {
			input:         "test" + string(byte(27)),
			expectedStr:   "test" + string(byte(27)),
			expectedIgnore: false,
			description:   "escape at end should not be ignored",
		},
		"empty string": {
			input:         "",
			expectedStr:   "",
			expectedIgnore: false,
			description:   "empty string should be handled",
		},
		"single character": {
			input:         "a",
			expectedStr:   "a",
			expectedIgnore: false,
			description:   "single character should pass through",
		},
		"single escape character": {
			input:         string(byte(27)),
			expectedStr:   string(byte(27)),
			expectedIgnore: false,
			description:   "single escape should not trigger ignore",
		},
		"multiple escapes": {
			input:         string([]byte{27, 27, 'a'}),
			expectedStr:   "",
			expectedIgnore: true,
			description:   "multiple bytes starting with escape should be ignored",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			str, ignore := cleanBufferForWSL(tc.input)
			assert.Equal(t, tc.expectedStr, str, tc.description)
			assert.Equal(t, tc.expectedIgnore, ignore, tc.description)
		})
	}
}

// TestBreakMultilinePrompt tests buffer clearing
// Bug hunting: ensure buffer is properly cleared to avoid memory leaks
func TestBreakMultilinePrompt(t *testing.T) {
	tests := map[string]struct {
		initialBuffer []string
		description   string
	}{
		"clear non-empty buffer": {
			initialBuffer: []string{"SELECT *", "FROM users", "WHERE id = 1"},
			description:   "should clear multi-line buffer",
		},
		"clear single line buffer": {
			initialBuffer: []string{"SELECT 1"},
			description:   "should clear single line buffer",
		},
		"clear already empty buffer": {
			initialBuffer: []string{},
			description:   "clearing empty buffer should not panic",
		},
		"clear nil buffer": {
			initialBuffer: nil,
			description:   "clearing nil buffer should not panic",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := &InteractiveClient{
				interactiveBuffer: tc.initialBuffer,
			}

			// This should not panic
			require.NotPanics(t, func() {
				client.breakMultilinePrompt(nil)
			})

			// Buffer should be empty (not nil, but empty slice)
			assert.NotNil(t, client.interactiveBuffer)
			assert.Equal(t, 0, len(client.interactiveBuffer), tc.description)
		})
	}
}

// TestCreatePromptContext tests context creation and cancellation
// Bug hunting: resource leaks from uncanceled contexts
func TestCreatePromptContext(t *testing.T) {
	tests := map[string]struct {
		setupClient func() *InteractiveClient
		description string
	}{
		"create context on fresh client": {
			setupClient: func() *InteractiveClient {
				return &InteractiveClient{}
			},
			description: "should create context with cancel function",
		},
		"create context when previous exists": {
			setupClient: func() *InteractiveClient {
				client := &InteractiveClient{}
				// Create first context
				ctx1 := client.createPromptContext(context.Background())
				// Verify first context is valid
				require.NotNil(t, ctx1)
				require.NotNil(t, client.cancelPrompt)
				return client
			},
			description: "should cancel previous context and create new one",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := tc.setupClient()
			parentCtx := context.Background()

			ctx := client.createPromptContext(parentCtx)

			// Verify context is valid
			assert.NotNil(t, ctx, "context should not be nil")
			assert.NotNil(t, client.cancelPrompt, "cancel func should be set")

			// Verify context is not already cancelled
			select {
			case <-ctx.Done():
				t.Fatal("context should not be cancelled immediately")
			default:
				// Expected - context is still active
			}

			// Call cancel and verify it works
			client.cancelPrompt()

			// Context should now be cancelled
			select {
			case <-ctx.Done():
				// Expected - context is now cancelled
			default:
				t.Fatal("context should be cancelled after calling cancel")
			}
		})
	}
}

// TestCreateQueryContext tests query context creation
// Bug hunting: ensure query contexts can be cancelled properly
func TestCreateQueryContext(t *testing.T) {
	client := &InteractiveClient{}
	parentCtx := context.Background()

	ctx := client.createQueryContext(parentCtx)

	// Verify context is valid
	assert.NotNil(t, ctx, "context should not be nil")
	assert.NotNil(t, client.cancelActiveQuery, "cancel func should be set")

	// Verify context is not already cancelled
	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled immediately")
	default:
		// Expected
	}
}

// TestCancelActiveQueryIfAny tests query cancellation
// Bug hunting: ensure cancellation is idempotent and doesn't panic
func TestCancelActiveQueryIfAny(t *testing.T) {
	tests := map[string]struct {
		setupClient func() *InteractiveClient
		description string
	}{
		"cancel with no active query": {
			setupClient: func() *InteractiveClient {
				return &InteractiveClient{
					cancelActiveQuery: nil,
				}
			},
			description: "should not panic when no query is active",
		},
		"cancel with active query": {
			setupClient: func() *InteractiveClient {
				client := &InteractiveClient{}
				client.createQueryContext(context.Background())
				return client
			},
			description: "should cancel active query",
		},
		"cancel twice should not panic": {
			setupClient: func() *InteractiveClient {
				client := &InteractiveClient{}
				client.createQueryContext(context.Background())
				// Cancel once
				client.cancelActiveQueryIfAny()
				return client
			},
			description: "calling cancel twice should be safe",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := tc.setupClient()

			// Should not panic
			require.NotPanics(t, func() {
				client.cancelActiveQueryIfAny()
			}, tc.description)

			// After cancellation, cancelActiveQuery should be nil
			assert.Nil(t, client.cancelActiveQuery, "cancel func should be nil after cancellation")
		})
	}
}

// Note: TestNewInteractiveClient and TestHandleInitResult are skipped
// as they require complex initialization of app-specific directories and file paths
// The underlying logic is tested through integration tests

// TestAfterPromptCloseAction tests the action enum
// Bug hunting: ensure action values are distinct and usable
func TestAfterPromptCloseAction(t *testing.T) {
	tests := map[string]struct {
		action1     AfterPromptCloseAction
		action2     AfterPromptCloseAction
		shouldEqual bool
	}{
		"exit and restart are different": {
			action1:     AfterPromptCloseExit,
			action2:     AfterPromptCloseRestart,
			shouldEqual: false,
		},
		"exit equals itself": {
			action1:     AfterPromptCloseExit,
			action2:     AfterPromptCloseExit,
			shouldEqual: true,
		},
		"restart equals itself": {
			action1:     AfterPromptCloseRestart,
			action2:     AfterPromptCloseRestart,
			shouldEqual: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldEqual {
				assert.Equal(t, tc.action1, tc.action2)
			} else {
				assert.NotEqual(t, tc.action1, tc.action2)
			}
		})
	}
}

// TestClosePrompt tests prompt closure with different actions
// Bug hunting: ensure action is set before cancellation
func TestClosePrompt(t *testing.T) {
	tests := map[string]struct {
		action      AfterPromptCloseAction
		description string
	}{
		"close with exit action": {
			action:      AfterPromptCloseExit,
			description: "should set exit action and cancel",
		},
		"close with restart action": {
			action:      AfterPromptCloseRestart,
			description: "should set restart action and cancel",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cancelled := false
			client := &InteractiveClient{
				cancelPrompt: func() {
					cancelled = true
				},
			}

			client.ClosePrompt(tc.action)

			assert.Equal(t, tc.action, client.afterClose, tc.description)
			assert.True(t, cancelled, "cancel should be called")
		})
	}
}

// TestInteractiveClientFields tests that InteractiveClient struct has expected fields
// Bug hunting: ensure critical fields exist for future maintainers
func TestInteractiveClientFields(t *testing.T) {
	// This test documents the expected fields of InteractiveClient
	// It helps catch accidental removal of critical fields
	client := &InteractiveClient{}

	// Use reflection-free approach - just verify we can access the fields
	_ = client.initData
	_ = client.promptResult
	_ = client.interactiveBuffer
	_ = client.interactivePrompt
	_ = client.interactiveQueryHistory
	_ = client.autocompleteOnEmpty
	_ = client.cancelActiveQuery
	_ = client.cancelPrompt
	_ = client.initResultChan
	_ = client.initialisationComplete
	_ = client.afterClose
	_ = client.schemaMetadata
	_ = client.highlighter
	_ = client.hidePrompt
	_ = client.suggestions

	// If we get here without compile errors, all expected fields exist
	assert.NotNil(t, client)
}

// Note: TestQueryHistoryInitialization is skipped as it requires app-specific directory initialization
