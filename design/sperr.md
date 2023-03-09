# New package `sperr`

## `sperr.Error`

An `sperr.Error` is a stateful object with a `StackTrace` till the point of creation with a stack depth of `32` (`32` picked OTA)

`sperr.Error` satisfies the standard `error` interface.

## Create `sperr.Error`:

> **Note:** All `sperr.Error` factory functions return an `error` interface.

### `sperr.New(format string, args interface{}...)`

This is to be used when we want to create new `error` instances. Always carries a `StackTrace`. It is recommended that this function be called from the actual place of the error and not to create error.

### `sperr.Wrap(err error, options Option...)`

If the given `err` is not an `sperr.Error`, this wraps around `err` and creates an `sperr.Error` along with a `StackTrace`.

Returns `nil` if `err` is `nil`. `Wrap` tries to infer a friendly message for the error and if the inference succeeded, it will set the friendly message as it's own message.

### `sperr.WrapWithMessage(err error, format string, args ...interface{})`

Wrap an `error` to create an `sperr.Error` and sets a formatted message to the `wrapper`.

`WrapWithMessage` is functionally equivalent to `Wrap(err, WithMessage(format,args...))` - but maintains the proper call stack.

### `sperr.WrapWithRootMessage(err error, format string, args ...interface{})`

Wrap an `error` to create an `sperr.Error` and sets a formatted message to the `wrapper` along with the `root` flag.

`WrapWithRootMessage` is functionally equivalent to `Wrap(err, WithRootMessage(format,args...))` - but maintains the proper call stack.

### `sperr.ToError(val interface{})`

This creates an `error` object from any available value.

If `val` is an instance of `error`, `ToError` creates a wrapper around `val` and returns it. Otherwise, it creates a new error using the value of `fmt.Sprintf("%v", val)` as the message. In both the cases, `ToError` creates and includes a `StackTrace`.

## Adding Options

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

## Printing errors

`sperr.Error` objects implement the `Formatter` interface to facilitate serializing errors to `io.Writer` interfaces.

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
    return sperr.WrapWithRootMessage(err, "could not open file at %s", path)
  }
  return nil
}

func wrapWithMessageAndDetail() error {
  err := readFile()

  return sperr.Wrap(
    err,
    sperr.WithMessage("message from wrapWithMessageAndDetail"),
    sperr.WithDetail("detail from wrapWithMessageAndDetail"),
  )
}

showCaseErr := sperr.Wrap(
  err,
  sperr.WithMessage("message from main"),
  sperr.WithDetail("detail from main"),
)

```

Outputs of the `showCaseErr` in preceeding program would be:

#### `%q`

`"message from main : message from wrapWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path"`

#### `%s` and `%v`

`message from main : message from wrapWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path`

#### `%+v`

```
message from main : message from wrapWithMessageAndDetail : could not open file at /imaginary/path

Details:
message from main :: detail from main
|-- message from wrapWithMessageAndDetail :: detail from wrapWithMessageAndDetail
|-- could not open file at /imaginary/path
|-- open /imaginary/path: no such file or directory
```

#### `%#v`

```
message from main : message from wrapWithMessageAndDetail : could not open file at /imaginary/path : open /imaginary/path

Details:
message from main :: detail from main
|-- message from wrapWithMessageAndDetail :: detail from wrapWithMessageAndDetail
|-- could not open file at /imaginary/path : open /imaginary/path: no such file or directory

Stack:
main.readFile
        /home/user/sandbox/main.go:83
main.wrapWithMessageAndDetail
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

### Create `error` with `message` and `detail`

```
func validateData(data int) error {
  if data > 10 {
    return sperr.Wrap(
      sperr.New("invalid argument: %d", data),
      sperr.WithDetail("error occurred with %d argument", data),
    )
  }
  return nil
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
  return nil, sperr.WrapWithMessage(err, "error unmarshalling file content in %s", filePath)
}
```

or

```
if err := json.Unmarshal(byteContent, &data); err != nil {
  return nil, sperr.Wrap(err, sperr.WithMessage("error unmarshalling file content in %s", filePath))
}
```

### Wrap an `error` with `detail`

```
err := validateData(userInput.numAttacks)
if err!= nil {
  return sperr.Wrap(err, sperr.WithDetail("error occurred with %d argument", userInput.numAttacks))
}
```

### Wrap an `error` with a message replacing the message of the original `error`

```
if _, err := installFDW(ctx, false); err != nil {
	log.Printf("[TRACE] installFDW failed: %v", err)
	return sperr.WrapWithRootMessage(err, "Update steampipe-postgres-fdw... FAILED!")
}
```

or

```
if _, err := installFDW(ctx, false); err != nil {
	log.Printf("[TRACE] installFDW failed: %v", err)
	return sperr.Wrap(err, sperr.WithRootMessage("Update steampipe-postgres-fdw... FAILED!"))
}
```

> Setting an error as the `root` error hides all errors below it from the user interface. They are not purged - just hidden from display when displaying error messages. When enumerating error `details`, the details of all errors in the stack are shown - including errors under a `root` error.

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

#### `WrapWithMessage`

The package function `WrapWithMessage` **always** wraps around the `error` given to it. This is because `WrapWithMessage` always sets it's own message with the arguments provided.

#### `WithMessage`

`WithMessage` sets the internal `message` if it is empty. Otherwise, it will create a `wrapper` around it's instance and set the `message` on the `wrapper` and returns the `wrapper`. This ensures that `WithMessage` is never lossy - but only creates wrappers when necessary.

#### `WithDetail`

`WithDetail` behaves just like `WithMessage`, but on the `detail` property.

#### Example:

> ```
> sperr.WrapWithMessage(
>   sperr.Wrap(
>     err,
>     sperr.WithDetail("added detail"),
>     sperr.WithMessage("error occurred with %d argument", intArgument),
>   ),
>   "error occurred"
> )
> ```
>
> Result:
>
> ```
> Error {
>   Error {
>     err
>     Message : "error occurred with 10 argument"
>     Detail  : "added detail"
>   }
>   Message : "error occurred"
> }
> ```
