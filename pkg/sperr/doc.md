New package `sperr`

# `sperr.Error`

`sperr.Error` satisfies the standard `error` interface. An `sperr.Error` is a stateful object with a StackTrace of the call stack to a depth of `32` (`32` picked OTA)

## Options

### `WithMessage(format string, args ...interface())`

Sets the formatted string to `error` if the `message` property is `empty`. Otherwise creates a new `error` by wrapping around this `error` and sets the message on the `wrapper`.

### `WithDetail(format string, args ...interface())`

Sets the formatted string to the `error` if the `detail` property is `empty`. Otherwise creates a new `error` by wrapping around this `error` and sets the detail on the `wrapper`.

### `WithRootMessage(format string, args ...interface())`

Sets the given formatted string as the error message and hides all error under this error from the UI. Setting the message follows the same rules as `WithMessage`. The `root` flag is set on the `error` returned by `WithMessage`.

### Using Options

```
sperr.Wrap(
  err,
  sperr.WithMessage("operation '%s' failed", operation),
  sperr.WithDetail("argument: %d", input),
)
```

## Create `sperr.Error`:

### `sperr.New(format string, args interface{}...)`

This is to be used when we want to create new `error` instances. Always carries a `StackTrace`. It is recommended that this function be called from the actual place of the error and not to create error.

### `sperr.Wrap(err error, options Option...)`

If the given `err` is not an `sperr.Error`, this wraps around `err` and creates an `sperr.Error` along with a `StackTrace`. Returns `nil` if `err` is `nil`. `Wrap` tries to infer a friendly message for the error and if the inference succeeded, it will set the friendly message as it's own message.

### `sperr.WrapWithMessage(err error, format string, args ...interface{})`

Wrap an `error` to create an `sperr.Error` and sets a formatted message to the `wrapper`.

`WrapWithMessage` is functionally equivalent to `Wrap(err, WithMessage(format,args...))` - but maintains the proper call stack.

### `sperr.WrapWithRootMessage(err error, format string, args ...interface{})`

Wrap an `error` to create an `sperr.Error` and sets a formatted message to the `wrapper` along with the `root` flag.

`WrapWithRootMessage` is functionally equivalent to `Wrap(err, WithRootMessage(format,args...))` - but maintains the proper call stack.

## Printing errors

`sperr.Error` objects implement the `Formatter` interface to facilitate serializing errors to output channels.

Formatting verbs supported are:
| | |
|-----|----------|
|`%s` | Print the error string |
|`%v` | See `%s` |
|`%+v`| `%v` along with the `detail` and `message` values of all the errors |
|`%#v`| `%+v` along the stacktrace of the underlying leaf error. Overrides `%+v`. |
|`%q` | Print the error string - double quoted and safely escaped with Go syntax |

### Example

Let's write up a minimal example program:

```
func readFile() error {
	path := "/imaginary/path"
	_, err := os.Open(path)
	if err != nil {
		return sperr.WrapWithMessage(err, "could not open file at %s", path)
	}
	return nil
}

func wrapFirstWithMessageAndDetail() error {
	err := readFile()

	return sperr.Wrap(
		err,
		sperr.WithMessage("message from wrapFirstWithMessageAndDetail"),
		sperr.WithDetail("detail from wrapFirstWithMessageAndDetail"),
	)
}

showCaseErr := sperr.Wrap(
  err,
  sperr.WithMessage("message from main"),
  sperr.WithDetail("detail from main"),
)

```

Outputs of the `showCaseErr` in preceeding program would be:

#### `%s`

`message from main : message from wrapFirstWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path: no such file or directory`

#### `%q`

`"message from main : message from wrapFirstWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path: no such file or directory"`

#### `%v`

`message from main : message from wrapFirstWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path: no such file or directory`

#### `%+v`

```
message from main : message from wrapFirstWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path: no such file or directory

Details:
message from main :: detail from main
|-- message from wrapFirstWithMessageAndDetail :: detail from wrapFirstWithMessageAndDetail
|-- could not open file at /imaginary/path : open /imaginary/path: no such file or directory
```

#### `%#v`

```
message from main : message from wrapFirstWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path: no such file or directory

Details:
message from main :: detail from main
|-- message from wrapFirstWithMessageAndDetail :: detail from wrapFirstWithMessageAndDetail
|-- could not open file at /imaginary/path : open /imaginary/path: no such file or directory

Stack:
main.readFile
        /home/user/sandbox/main.go:83
main.wrapFirstWithMessageAndDetail
        /home/user/sandbox/main.go:63
main.addMsgAndDetailToError
        /home/user/sandbox/main.go:53
main.wrapErrorAndSetRootMessage
        /home/user/sandbox/main.go:39
main.main
        /home/user/sandbox/main.go:33
runtime.main
        /usr/local/go/src/runtime/proc.go:250
runtime.goexit
        /usr/local/go/src/runtime/asm_arm64.s:1165
```

> Note: `%+#v` is functionally equivalent to `%#v`

## Examples:

Snippets from Steampipe code base:

### Create a new `error`

```
dbState, err := GetState()
if err != nil {
  log.Println("[TRACE] Error while loading database state", err)
  return err
}
if dbState != nil {
  return sperr.New("cannot install db - a previous version of the Steampipe service is still running. To stop running services, use %s ", constants.Bold("steampipe service stop"))
}
```

### Wrap an `error`

```
if err := json.Unmarshal(bytContent, &data); err != nil {
  return nil, sperr.Wrap(err)
}
```

### Wrap an `error` with a `message`

```
if err := json.Unmarshal(byteContent, &data); err != nil {
  return nil, sperr.Wrapf(err, "error unmarshalling file content in %s", filePath)
}
```

or

```
if err := json.Unmarshal(byteContent, &data); err != nil {
  return nil, sperr.Wrap(err).WithMessage("error unmarshalling file content in %s", filePath)
}
```

### Create an `error` with `message` and `detail`

```
func validateData(data int) error {
  if data > 10 {
    return sperr.New("invalid argument: %d", data).WithDetail("error occurred with %d argument", data)
  }
  return nil
}
```

### Wrap an `error` with `detail`

```
err := validateData(userInput.numAttacks)
if err!= nil {
  return sperr.Wrap(err).WithDetail("error occurred with %d argument", userInput.numAttacks)
}
```

### Wrap an `error` with `message`

```
err := validateData(userInput.numAttacks)
if err!= nil {
  return sperr.Wrap(err).WithMessage("error occurred with %d argument", userInput.numAttacks)
}
```

> While wrapping around `err`, if `Wrap` could infer a `message` then `WithMessage` will create a `wrapper` around the **output of `sperr.Wrap(err)`** and set the message on the `wrapper`. Otherwise, it will just set the `message` on `err`.

### Wrap an `error` with a message replacing the message of the `error`

```
err = sperr.Wrap(err).AsRootMessage()
```

```
if _, err := installFDW(ctx, false); err != nil {
	log.Printf("[TRACE] installFDW failed: %v", err)
	return sperr.Wrapf(err, "Update steampipe-postgres-fdw... FAILED!").AsRootMessage()
}
```

> Setting an error as the `root` error hides all errors below it from the user interface. They are not purged - just hidden from display when displaying error messages. When enumerating error `details`, the details of all errors in the stack are shown - including errors under a `root` error.

### Wrap an `error` with formatted message and then set a `message`

```
err = sperr.Wrapf(err, "error occurred").WithMessage("error occurred with %d argument", intArgument)
```

### Convert `panic` recovery to an `error`

```
defer func() {
  if r := recover(); r != nil {
    err = sperr.ToError(r)
  }
}()
```

## Technicalities

### Wrapping as necessary

#### `Wrap`

The package function `Wrap` wraps around a given `error` instance if and only if it is not an instance of `sperr.Error`. This effectively ensures that the return of `Wrap` is always an instance of `sperr.Error`.

#### `Wrapf`

The package function `Wrapf` **always** wraps around the `error` given to it. This is because `Wrapf` always sets it's own message with the arguments provided.

#### `WithMessage`

`WithMessage` sets the internal `message` if it is empty. Otherwise, it will create a `wrapper` around it's instance and set the `message` on the `wrapper` and returns the `wrapper`. This ensures that `WithMessage` is never lossy - but only creates wrappers when necessary.

#### `WithDetail`

`WithDetail` behaves just like `WithMessage`, but on the `detail` property.

#### Example:

> ```
> err = sperr.Wrapf(err, "error occurred").
>              WithDetail("added detail").
>              WithMessage("error occurred with %d argument", intArgument)
> ```
>
> Result:
>
> ```
> Error {
>   Error {
>     err
>     Message : "error occurred"
>     Detail  : "added detail"
>   }
>   Message : "error occurred with 10 argument"
> }
> ```
